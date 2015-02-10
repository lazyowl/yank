package fileFetcher

import (
	"yank/network"
	"yank/fileManager"
	"yank/config"
	"yank/cache"
	"yank/constants"
	"fmt"
	"time"
)

const (
	UNREQUESTED = 0
	REQUEST_SENT = 1
	REQUEST_COMPLETE = 2
)

var fileController *fileManager.FileController
var fileListCache *cache.UserFileCache
var peer *network.Peer
var hostCache *cache.HostCache

var positionMap map[int]int
var userFetchMap map[string][]int
var userTimeout map[string]int
var potentialUserMap map[string]fileManager.MyFile
var timer *time.Ticker
var numOutstandingRequests int
var localFile *fileManager.MyFile
var currentFileRequestHash string

var downloadInProgress bool


type FileFetcher struct {
	ClientQ chan string		// queue of file requests made by client
	ResponseQ chan network.CmdMessage	// queue of file request responses by other peers
	ServerQ chan network.CmdMessage	// queue of file requests made by other peers on the network
}

func NewFileFetcher(fc *fileManager.FileController, p *network.Peer, hc *cache.HostCache, fcache *cache.UserFileCache) *FileFetcher {
	ff := FileFetcher{}
	// TODO probably make these buffered
	ff.ClientQ = make(chan string)
	ff.ResponseQ = make(chan network.CmdMessage)
	ff.ServerQ = make(chan network.CmdMessage)

	positionMap = make(map[int]int)
	userFetchMap = make(map[string][]int)
	userTimeout = make(map[string]int)
	potentialUserMap = make(map[string]fileManager.MyFile)

	fileController = fc
	peer = p
	hostCache = hc
	fileListCache = fcache
	timer = time.NewTicker(time.Second)
	numOutstandingRequests = 0

	return &ff
}

// starts sending round and returns the number of requests that were sent
// return value of 0 indicates nothing more can be done
func StartSendingRound() {
	for k, v := range potentialUserMap {
		if numOutstandingRequests > constants.MAX_FILE_REQUESTS {
			break
		}
		chunks := []int{}
		for i := 0; i < v.NumBlocks(); i++ {
			if (localFile != nil && localFile.HashBitVector.GetBit(uint(i)) == false) || localFile == nil {
				val, found := positionMap[i];
				if (found && val == UNREQUESTED) || !found {
					// you may request
					chunks = append(chunks, i)
					positionMap[i] = REQUEST_SENT
				}
			}
		}

		if len(chunks) == 0 {
			continue
		}

		userFetchMap[k] = chunks
		userTimeout[k] = constants.REQUEST_TTL

		m := network.NewCmdMessage()
		m.Source = config.Config.Name
		m.Cmd = constants.FILE_REQUEST
		m.Hash = currentFileRequestHash
		m.RequestedChunkNumbers = chunks
		peer.SendUnicast(m.Serialize(), hostCache.Get(k))
		numOutstandingRequests++
	}
}

func (ff *FileFetcher) ManageFileFetch() {
	for{
		select {
			case <-timer.C: {
				if !downloadInProgress {
					break
				}
				// tick
				somethingTimedOut := false
				for k, _ := range userTimeout {
					userTimeout[k]--
					if userTimeout[k] == 0 {
						numOutstandingRequests--
						somethingTimedOut = true

						// remove this user from the potentialList
						delete(potentialUserMap, k)

						// reset the status of all the positions that that user was going to return
						for _, pos := range userFetchMap[k] {
							positionMap[pos] = UNREQUESTED
						}
					}
				}
				if somethingTimedOut {
					StartSendingRound()
				}
			}
			case fileResponse := <-ff.ResponseQ: {
				// if this is a delayed reply, ignore
				if !downloadInProgress {
					break
				}

				// if a new file is being downloaded, ignore
				if currentFileRequestHash != fileResponse.Hash {
					break
				}

				// this is a stale reply. the timer for this source already timed out. ignore this
				if _, found := userFetchMap[fileResponse.Source]; !found {
					break
				}
				numOutstandingRequests--
				if localFile == nil {
					// create file (for now, with the hash as the name), ideally, this would be specified by the user TODO
					localFile, _ = fileController.CreateEmptyFile(currentFileRequestHash, currentFileRequestHash, fileResponse.Size)
				}
				localFile.Open()
				for _, tuple := range fileResponse.ReturnedDataChunks {
					err := localFile.WriteChunk(tuple.Position, tuple.Data)
					if err != nil {
						fmt.Println("Write Err:", err)
						continue
					}
					positionMap[tuple.Position] = REQUEST_COMPLETE
					localFile.HashBitVector.SetBit(uint(tuple.Position))
				}
				localFile.Close()
				StartSendingRound()
			}
			case fileRequestHash := <-ff.ClientQ: {
				// potentialUserMap = map of user -> file
				currentFileRequestHash = fileRequestHash
				potentialUserMap = fileListCache.GetExistingByHash(fileRequestHash)
				if len(potentialUserMap) == 0 {
					fmt.Println("Nobody has it!")
					break
				}
				localFile = fileController.FileFromHash(fileRequestHash)

				// download has begun
				downloadInProgress = true

				StartSendingRound()
			}
			case cmdMsg := <-ff.ServerQ: {
				hash := cmdMsg.Hash
				f := fileController.FileFromHash(hash)
				if f == nil {
					// don't send any response, let my timer on the sender's side time out
					break
				}
				data := []network.DataTuple{}
				f.Open()
				for _, pos := range cmdMsg.RequestedChunkNumbers {
					b, err := f.ReadChunk(pos)
					if err != nil {
						continue
					} else {
						data = append(data, network.NewDataTuple(pos, b))
					}
				}
				f.Close()

				response := network.NewCmdMessage()
				response.Cmd = constants.FILE_RESPONSE
				response.Source = config.Config.Name
				response.Hash = hash
				response.ReturnedDataChunks = data
				response.Size = f.Size
				peer.SendUnicast(response.Serialize(), hostCache.Get(cmdMsg.Source))
			}
		}
	}
}
