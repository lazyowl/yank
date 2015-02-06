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

type CmdMessage struct {
	// control messaging
	Cmd int
	Files []fileManager.MyFile
	Source string

	// data messaging
	Hash string
	RequestedChunkNumbers []int
	ReturnedDataChunks map[int][]byte
	Size int
}
func (m CmdMessage) Serialize() []byte {
	b, _ := json.Marshal(m)
	return b
}
func Deserialize(b []byte) CmdMessage {
	var msg CmdMessage
	json.Unmarshal(b, &msg)
	return msg
}
func NewCmdMessage() CmdMessage {
	cmdMsg := CmdMessage{}
	return cmdMsg
}


type FileFetcher struct {
	fileQueue chan string		// queue of hashes
	numRequests int
	RecvQ chan CmdMessage
	acceptNewRequests chan bool
}

func NewFileFetcher() *FileFetcher {
	ff := FileFetcher{}
	ff.fileQueue = make(chan string)
	ff.numRequests = 0
	ff.acceptNewRequests = make(chan bool)
	ff.RecvQ = make(chan CmdMessage)
	return &ff
}

func (ff *FileFetcher) EnqueueFileRequest(hash string) {
	ff.fileQueue <- hash
}

func (ff *FileFetcher) ManageFileFetch() {
	select {
		case fileResponse := <-ff.RecvQ: {
			hash := fileResponse.Hash
			f := fileController.FileFromHash(hash)
			if f == nil {
				// create file (for now, with the hash as the name), ideally, this would be specified by the user TODO
				f, _ = fileController.CreateEmptyFile(hash, hash, fileResponse.Size)
			}
			f.Open()
			for pos, dat := range fileResponse.ReturnedDataChunks {
				f.WriteChunk(pos, dat)
			}
			f.Close()
		}
		case <-ff.acceptNewRequests: {
			// handle new request
		}
	}
}


// ping to let everyone know that we are here and what we have
func ping(name string) {
	//m := CmdMessage{LIST_REPLY, fileController.ListLocalFiles(), config.Config.Name}
	m := NewCmdMessage()
	m.Cmd = LIST_REPLY
	m.Files = fileController.ListLocalFiles()
	m.Source = config.Config.Name
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
					cmdMsg := Deserialize(peerMsg.Msg.Value)
					hostCache.Put(cmdMsg.Source, peerMsg.From)
					switch cmdMsg.Cmd {
						case LIST: {
							response := NewCmdMessage()
							response.Cmd = LIST_REPLY
							response.Files = fileController.ListLocalFiles()
							response.Source = config.Config.Name
							msg := network.CreateMessage(response.Serialize())
							peer.SendUnicast(msg, peerMsg.From)
						}
						case LIST_REPLY: {
							// possible TODO: maybe just send the deltas each time?
							fileListCache.ClearUser(cmdMsg.Source)
							for _, f := range cmdMsg.Files {
								fileListCache.Put(cmdMsg.Source, f)
							}
						}
						case FILE_REQUEST: {
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
