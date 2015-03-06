package fileFetcher

import (
	"yank/network"
	"yank/fileManager"
	"yank/config"
	"yank/cache"
	"fmt"
	"time"
)

const (
	UNREQUESTED = iota
	REQUEST_SENT
	REQUEST_COMPLETE
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
var fileSaveName string
var currentFileRequestHash string

var downloadInProgress bool

type FileToFetch struct {
	Hash string
	Name string
}

type FileFetcher struct {
	ClientQ chan FileToFetch		// queue of file requests made by client
	ResponseQ chan network.CmdMessage	// queue of file request responses by other peers
	ServerQ chan network.CmdMessage	// queue of file requests made by other peers on the network
	DownloadComplete chan bool		// notifies app that download is complete
}

func NewFileFetcher(fc *fileManager.FileController, p *network.Peer, hc *cache.HostCache, fcache *cache.UserFileCache) *FileFetcher {
	ff := FileFetcher{}
	// TODO probably make these buffered
	ff.ClientQ = make(chan FileToFetch)
	ff.ResponseQ = make(chan network.CmdMessage)
	ff.ServerQ = make(chan network.CmdMessage)
	ff.DownloadComplete = make(chan bool)

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

// starts sending round
// IMPORTANT: if after the sending round, the number of outstanding requests is zero,
// I say that the download is complete
func StartSendingRound() {
	for k, v := range potentialUserMap {
		if numOutstandingRequests > config.Config.MaxFileRequests {
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
		userTimeout[k] = config.Config.RequestTTL

		m := network.NewCmdMessage()
		m.Source = config.Config.Name
		m.Cmd = network.FILE_REQUEST
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
				if somethingTimedOut && downloadInProgress {
					StartSendingRound()
					if numOutstandingRequests == 0 {
						downloadInProgress = false
						ff.onComplete()
					}
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
					localFile, _ = fileController.CreateEmptyFile(fileSaveName, currentFileRequestHash, fileResponse.Size)
				}
				localFile.Open()
				for _, tuple := range fileResponse.ReturnedDataChunks {
					err := localFile.WriteChunk(tuple.Position, tuple.Data, tuple.Size)
					if err != nil {
						fmt.Println("Write Err:", err)
						continue
					}
					positionMap[tuple.Position] = REQUEST_COMPLETE
					localFile.HashBitVector.SetBit(uint(tuple.Position))
				}
				localFile.Close()
				if downloadInProgress {
					StartSendingRound()
					if numOutstandingRequests == 0 {
						downloadInProgress = false
						ff.onComplete()
					}
				}
			}
			case fileToFetch := <-ff.ClientQ: {
				// potentialUserMap = map of user -> file
				fileRequestHash := fileToFetch.Hash
				if fileToFetch.Name != "" {
					fileSaveName = fileToFetch.Name
				} else {
					fileSaveName = fileRequestHash
				}
				currentFileRequestHash = fileRequestHash
				potentialUserMap = fileListCache.GetExistingByHash(fileRequestHash)
				if len(potentialUserMap) == 0 {
					fmt.Println("Nobody has it!")
					ff.DownloadComplete <- false
					break
				}
				localFile = fileController.FileFromHash(fileRequestHash)

				// download has begun
				downloadInProgress = true

				if downloadInProgress {
					StartSendingRound()
					if numOutstandingRequests == 0 {
						downloadInProgress = false
						ff.onComplete()
					}
				}
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
					b, size, err := f.ReadChunk(pos)
					if err != nil {
						continue
					} else {
						data = append(data, network.NewDataTuple(pos, b, size))
					}
				}
				f.Close()

				response := network.NewCmdMessage()
				response.Cmd = network.FILE_RESPONSE
				response.Source = config.Config.Name
				response.Hash = hash
				response.ReturnedDataChunks = data
				response.Size = f.Size
				peer.SendUnicast(response.Serialize(), hostCache.Get(cmdMsg.Source))
			}
		}
	}
}

func (ff *FileFetcher) onComplete() {
	fmt.Println("==== Download Complete ===")
	fmt.Printf("Name:%s\nHash:%s\nPercent Complete:%d%%\nTrue Size:%d\n", localFile.Name, localFile.FullHash, localFile.PercentComplete(), localFile.Size)
	ff.DownloadComplete <- true
}
