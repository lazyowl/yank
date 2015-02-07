// Package main contains the loops for listening and sending messages.
package main

import (
	"yank/config"
	"yank/network"
	"yank/fileManager"
	"yank/cache"
	"fmt"
	"strings"
	"bufio"
	"os"
	"time"
)

const (
	PING_INTERVAL = 10		// multicast every PING_INTERVAL seconds if something has changed
	LIST = 0
	LIST_REPLY = 1

	FILE_REQUEST = 2
	FILE_RESPONSE = 3

	MAX_FILE_REQUESTS = 20	// maximum outstanding requests
)

// ignoring errors for now
var (
	peer, _ = network.NewPeer()
	fileController = fileManager.NewFileController()
	hostCache = cache.NewHostcache()
	fileListCache = cache.NewUserFileCache()
	fileFetchManager = NewFileFetcher()
)

type FileFetcher struct {
	FileQ chan string		// queue of hashes
	numRequests int
	RecvQ chan network.CmdMessage
}

func NewFileFetcher() *FileFetcher {
	ff := FileFetcher{}
	ff.FileQ = make(chan string)
	ff.numRequests = 0
	ff.RecvQ = make(chan network.CmdMessage)
	return &ff
}

func (ff *FileFetcher) ManageFileFetch() {
	for{
		select {
			case fileResponse := <-ff.RecvQ: {
				hash := fileResponse.Hash
				f := fileController.FileFromHash(hash)
				if f == nil {
					// create file (for now, with the hash as the name), ideally, this would be specified by the user TODO
					var err error
					f, err = fileController.CreateEmptyFile(hash, hash, fileResponse.Size)
					if err != nil {
						fmt.Println("[CreateEmptyFile Err]", err)
					}
				}
				openErr := f.Open()
				fmt.Println(openErr)
				for _, tuple := range fileResponse.ReturnedDataChunks {
					err := f.WriteChunk(tuple.Position, tuple.Data)
					if err != nil {
						continue
					}
					f.HashBitVector.SetBit(uint(tuple.Position))
				}
				f.Close()
			}
			case fileRequestHash := <-ff.FileQ: {
				// ideally limit the number of requests we send out TODO
				potentialFiles := fileListCache.GetExistingByHash(fileRequestHash)
				if len(potentialFiles) == 0 {
					fmt.Println("Nobody has it!")
					break
				}
				localFile := fileController.FileFromHash(fileRequestHash)
				chunkList := []int{}
				if localFile == nil {
					// i don't have it => request all chunks
					for _, v := range potentialFiles {
						for i := 0; i < v.NumBlocks(); i++ {
							chunkList = append(chunkList, i)
						}
						break
					}
				} else {
					// identify the chunks we need to request
					localFile.HashBitVector.ResetIterator()
					for i := 0; i < localFile.NumBlocks(); i++ {
						val, err := localFile.HashBitVector.Next()
						if err != nil {
							break
						}
						if val {
							chunkList = append(chunkList, i)
						}
					}
				}
				// TODO
				// naive algorithm for now:
				// for each chunk, cycle through each potentialFile and see if it has it
				// if so, fetch, else next
				fetchMap := make(map[string][]int)
				for _, pos := range chunkList {
					for u, potential := range potentialFiles {
						if potential.HashBitVector.GetBit(uint(pos)) {
							if _, found := fetchMap[u]; !found {
								fetchMap[u] = []int{}
							}
							fetchMap[u] = append(fetchMap[u], pos)
							break
						}
					}
				}
				// we now have a map (user -> chunks to request)
				// so build the messages and fire them out
				// for now, bundle all requested chunk positions in one message
				// TODO make sure this doesn't get too big (probably want to split)
				for u, chunks := range fetchMap {
					m := network.NewCmdMessage()
					m.Cmd = FILE_REQUEST
					m.Hash = fileRequestHash
					m.RequestedChunkNumbers = chunks
					peer.SendUnicast(m.Serialize(), hostCache.Get(u))
				}
			}
		}
	}
}


// ping to let everyone know that we are here and what we have
func ping(name string) {
	//m := network.CmdMessage{LIST_REPLY, fileController.ListLocalFiles(), config.Config.Name}
	m := network.NewCmdMessage()
	m.Cmd = LIST_REPLY
	m.Files = fileController.ListLocalFiles()
	m.Source = config.Config.Name
	peer.SendMulticast(m.Serialize())
}

// main contains all the loops
func main() {
	// setup listeners
	go peer.ListenUnicast()
	go peer.ListenMulticast()

	go fileFetchManager.ManageFileFetch()

	// ping (ideally needs to repeat)
	ping(config.Config.Name)
	pingTicker := time.NewTicker(time.Second * PING_INTERVAL)

	// raw peer loop
	go func() {
		for {
			select {
				case peerMsg := <-peer.RecvCh: {
					cmdMsg := network.Deserialize(peerMsg.Msg)
					fmt.Println("received message!", cmdMsg)
					// TODO check with IP addresses as well
					if cmdMsg.Source == config.Config.Name {
						break
					}
					hostCache.Put(cmdMsg.Source, peerMsg.From)
					switch cmdMsg.Cmd {
						case LIST: {
							response := network.NewCmdMessage()
							response.Cmd = LIST_REPLY
							response.Files = fileController.ListLocalFiles()
							response.Source = config.Config.Name
							peer.SendUnicast(response.Serialize(), peerMsg.From)
						}
						case LIST_REPLY: {
							// possible TODO: maybe just send the deltas each time?
							fileListCache.ClearUser(cmdMsg.Source)
							for _, f := range cmdMsg.Files {
								fileListCache.Put(cmdMsg.Source, f)
							}
						}
						case FILE_REQUEST: {
							hash := cmdMsg.Hash
							f := fileController.FileFromHash(hash)
							if f == nil {
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
							response.Cmd = FILE_RESPONSE
							response.Source = config.Config.Name
							response.Hash = hash
							response.ReturnedDataChunks = data
							response.Size = f.Size
							peer.SendUnicast(response.Serialize(), peerMsg.From)
						}
						case FILE_RESPONSE: {
							fileFetchManager.RecvQ <- cmdMsg
						}
					}
				}
				case <-pingTicker.C: {
					// TODO perhaps only ping if anything has changed since the last ping
					ping(config.Config.Name)
				}
			}
		}
	}()

	// logic client loop
	// switch case based on the type of command received as input
	// this is possibly temporary
	for {
		fmt.Printf("$ ")
		bio := bufio.NewReader(os.Stdin)
		line, _, _:= bio.ReadLine()
		toks := strings.Split(string(line), " ")
		switch toks[0] {
			case "ls": {
				// hit the cache
				cachedList := fileListCache.GetAll()
				fmt.Println(cachedList)
			}
			case "get": {
				if len(toks) > 1 {
					fmt.Println(toks[1])
					fileFetchManager.FileQ <- toks[1]
				} else {
					fmt.Println("get what?")
				}
			}
			case "lls": {
				for _, f := range fileController.ListLocalFiles() {
					fmt.Println(f)
				}
			}
			case "lu": {
				for k, v := range hostCache.Cache() {
					fmt.Println(k, v)
				}
			}
			case "q": {
				return
			}
			default: {
				fmt.Println("Invalid command!")
			}
		}
	}
}
