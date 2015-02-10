// Package main contains the loops for listening and sending messages.
package main

import (
	"yank/constants"
	"yank/config"
	"yank/cache"
	"yank/network"
	"yank/fileManager"
	"yank/fileFetcher"
	"fmt"
	"strings"
	"bufio"
	"os"
	"time"
	"log"
)

var (
	peer *network.Peer
	fileController *fileManager.FileController
	hostCache *cache.HostCache
	fileListCache *cache.UserFileCache
	fileFetchManager *fileFetcher.FileFetcher
)

func init() {
	var err error
	peer, err = network.NewPeer()
	if err != nil {
		log.Fatal(err)
	}
	fileController = fileManager.NewFileController()
	hostCache = cache.NewHostCache()
	fileListCache = cache.NewUserFileCache()
	fileFetchManager = fileFetcher.NewFileFetcher(fileController, peer, hostCache, fileListCache)
}


// ping to let everyone know that we are here and what we have
func ping(name string) {
	m := network.NewCmdMessage()
	m.Cmd = constants.LIST_REPLY
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
	pingTicker := time.NewTicker(time.Second * constants.PING_INTERVAL)

	// raw peer loop
	go func() {
		for {
			select {
				case peerMsg := <-peer.RecvCh: {
					cmdMsg := network.Deserialize(peerMsg.Msg)
					fmt.Printf("received message: %v, |%s|\n", cmdMsg, cmdMsg.Source)
					// TODO check with IP addresses as well
					if cmdMsg.Source == config.Config.Name {
						break
					}
					hostCache.Put(cmdMsg.Source, peerMsg.From)
					switch cmdMsg.Cmd {
						case constants.LIST: {
							response := network.NewCmdMessage()
							response.Cmd = constants.LIST_REPLY
							response.Files = fileController.ListLocalFiles()
							response.Source = config.Config.Name
							peer.SendUnicast(response.Serialize(), peerMsg.From)
						}
						case constants.LIST_REPLY: {
							// possible TODO: maybe just send the deltas each time?
							fileListCache.ClearUser(cmdMsg.Source)
							for _, f := range cmdMsg.Files {
								fileListCache.Put(cmdMsg.Source, f)
							}
						}
						case constants.FILE_REQUEST: {
							fileFetchManager.ServerQ <- cmdMsg
						}
						case constants.FILE_RESPONSE: {
							fileFetchManager.ResponseQ <- cmdMsg
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
					fileFetchManager.ClientQ <- toks[1]
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
