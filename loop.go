// Package main contains the loops for listening and sending messages.
package main

import (
	"yank/config"
	"yank/network"
	"yank/fileManager"
	"yank/cache"
	"encoding/json"
	"fmt"
	"strings"
	"bufio"
	"os"
	"time"
)

const (
	PING_INTERVAL = 10		// multicast every PING_INTERVAL seconds
	LIST = 0
	LIST_REPLY = 1
)

// ignoring errors for now
var (
	peer, _ = network.NewPeer()
	fileController = fileManager.NewFileController()
	hostCache = cache.NewHostcache()
	fileListCache = cache.NewUserFileCache()
)

type HighMessage struct {
	Cmd int
	Files []fileManager.MyFile
	Source string
}

// convert the message into a string
func (m HighMessage) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}

func Deserialize(b []byte) HighMessage {
	var msg HighMessage
	json.Unmarshal(b, &msg)
	return msg
}

// ping to let everyone know that we are here and what we have
func ping(name string) {
	m := HighMessage{LIST_REPLY, fileController.ListLocalFiles(), config.Config.Name}
	peer.SendMulticast(network.CreateMessage(m.Serialize()))
}

// main contains all the loops
func main() {
	// setup listeners
	go peer.ListenUnicast()
	go peer.ListenMulticast()

	// ping (ideally needs to repeat)
	ping(config.Config.Name)
	pingTicker := time.NewTicker(time.Second * PING_INTERVAL)

	// raw peer loop
	go func() {
		for {
			select {
				case peerMsg := <-peer.RecvCh: {
					highMsg := Deserialize(peerMsg.Msg.Value)
					hostCache.Put(highMsg.Source, peerMsg.From)
					switch highMsg.Cmd {
						case LIST: {
							response := HighMessage{LIST_REPLY, fileController.ListLocalFiles(), config.Config.Name}
							msg := network.CreateMessage(response.Serialize())
							peer.SendUnicast(msg, peerMsg.From)
						}
						case LIST_REPLY: {
							// possible TODO: maybe just send the deltas each time?
							fileListCache.ClearUser(highMsg.Source)
							for _, f := range highMsg.Files {
								fileListCache.Put(highMsg.Source, f)
							}
						}
					}
				}
				case <-pingTicker.C: {
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
