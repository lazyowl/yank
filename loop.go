// Package main contains the loops for listening and sending messages.
package main

import (
	"yank/config"
	"yank/network"
	"yank/fileManager"
	"yank/cache"
	"yank/ui"
	"encoding/json"
	"fmt"
	"strings"
	"bufio"
	"os"
)

const (
	WAIT_DURATION = 2
	LIST = 0
	LIST_REPLY = 1
)

var (
	peer, _ = network.NewPeer()
	fileController = fileManager.NewFileController()
	hostCache = cache.NewHostcache()
	fileListCache = cache.NewFileListCache()
)

type HighMessage struct {
	Cmd int
	Files []*fileManager.MyFile
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
	m := HighMessage{LIST, nil, config.Config.Name}
	peer.SendMulticast(network.CreateMessage(m.Serialize()))
}

// main contains all the loops
func main() {
	// setup listeners
	go peer.ListenUnicast()
	go peer.ListenMulticast()

	// ping (ideally needs to repeat)
	ping(config.Config.Name)

	// raw peer loop
	go func() {
		for {
			peerMsg := <-peer.RecvCh
			highMsg := Deserialize(peerMsg.Msg.Value)
			hostCache.Put(highMsg.Source, peerMsg.From)
			switch highMsg.Cmd {
				case LIST: {
					response := HighMessage{LIST_REPLY, fileController.ListLocalFiles(), config.Config.Name}
					msg := network.CreateMessage(response.Serialize())
					peer.SendUnicast(msg, peerMsg.From)
				}
				case LIST_REPLY: {
					// add files to fileListCache
				}
			}
		}
	}()


	// logic client loop
	// switch case based on the type of command received as input
	for {
		fmt.Printf("$ ")
		bio := bufio.NewReader(os.Stdin)
		line, _, _:= bio.ReadLine()
		toks := strings.Split(string(line), " ")
		switch toks[0] {
			case "ls": {
				// hit the cache
			}
			case "get": {
				if len(toks) > 1 {
					fmt.Println(toks[1])
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
