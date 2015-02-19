// Package main contains the loops for listening and sending messages.
package main

import (
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
	"flag"
)

const DEFAULT_CONFIG_PATH = "./config.txt"

var (
	peer *network.Peer
	fileController *fileManager.FileController
	hostCache *cache.HostCache
	fileListCache *cache.UserFileCache
	fileFetchManager *fileFetcher.FileFetcher
)

func init() {
	var err error

	ifaceStr := flag.String("iface", "eth0", "LAN interface")
	configPathStr := flag.String("config", DEFAULT_CONFIG_PATH, "configuration file")
	nameStr := flag.String("name", "", "public name on the network")

	// these values cannot be <= 0, if they are <= 0, that means they weren't provided or the user entered bad values
	pingIntervalInt := flag.Int("ping", 0, "ping interval")
	maxRequestsInt := flag.Int("maxreq", 0, "max file requests")
	requestTTLInt := flag.Int("ttl", 0, "request TTL")

	flag.Parse();

	config.ReadConfig(*configPathStr, *nameStr, *pingIntervalInt, *maxRequestsInt, *requestTTLInt)

	peer, err = network.NewPeer(*ifaceStr)
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
	m.Cmd = network.LIST_REPLY
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
	pingTicker := time.NewTicker(time.Second * time.Duration(config.Config.PingInterval))

	// raw peer loop
	go func() {
		for {
			select {
				case peerMsg := <-peer.RecvCh: {
					cmdMsg := network.Deserialize(peerMsg.Msg)
					//fmt.Printf("received message: %v, |%s|\n", cmdMsg, cmdMsg.Source)
					// TODO check with IP addresses as well
					if cmdMsg.Source == config.Config.Name {
						break
					}
					hostCache.Put(cmdMsg.Source, peerMsg.From)
					switch cmdMsg.Cmd {
						case network.LIST: {
							response := network.NewCmdMessage()
							response.Cmd = network.LIST_REPLY
							response.Files = fileController.ListLocalFiles()
							response.Source = config.Config.Name
							peer.SendUnicast(response.Serialize(), peerMsg.From)
						}
						case network.LIST_REPLY: {
							// possible TODO: maybe just send the deltas each time?
							fileListCache.ClearUser(cmdMsg.Source)
							for _, f := range cmdMsg.Files {
								fileListCache.Put(cmdMsg.Source, f)
							}
						}
						case network.FILE_REQUEST: {
							fileFetchManager.ServerQ <- cmdMsg
						}
						case network.FILE_RESPONSE: {
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
		line, _, err:= bio.ReadLine()
		if err != nil {
			break
		}
		toks := strings.Split(string(line), " ")
		switch toks[0] {
			case "ls": {
				// hit the cache
				cachedList := fileListCache.GetAll()
				for u, m := range cachedList {
					fmt.Println("======", u, "======")
					for _, v := range m {
						fmt.Printf("%s\t%s\t%d\n", v.FullHash, v.Name, v.Size)
					}
				}
			}
			case "get": {
				if len(toks) == 2 {
					fileFetchManager.ClientQ <- fileFetcher.FileToFetch{toks[1], ""}
					<-fileFetchManager.DownloadComplete
				} else if len(toks) == 3 {
					fileFetchManager.ClientQ <- fileFetcher.FileToFetch{toks[1], toks[2]}
					<-fileFetchManager.DownloadComplete
				} else {
					fmt.Println("get what?")
				}
			}
			case "lls": {
				for _, f := range fileController.ListLocalFiles() {
					fmt.Printf("%s\t%s\t%d\n", f.FullHash, f.Name, f.Size)
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
