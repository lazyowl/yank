// Package main contains the loops for listening and sending messages.
package main

import (
	"fmt"
	"yank/network"
	"yank/config"
	"yank/fileManager"
	"yank/cache"
	"yank/util"
	"yank/ui"
	"encoding/json"
	"log"
	"time"
)

const (
	WAIT_DURATION = 2

	LIST = 0
	LIST_REPLY = 1
	PING = 2
	PING_REPLY = 3
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

// main contains all the loops
func main() {
	fmt.Println("Hello!")

	// read configuration
	errRead := config.ReadConfig()
	if errRead != nil {
		log.Fatal(errRead)
	}

	// peer
	peer, _ := network.NewPeer()
	go peer.ListenUnicast()
	go peer.ListenMulticast()

	// ping to let everyone know that we are here
	ping := HighMessage{PING, nil, config.Config.Name}
	msg := network.CreateMessage(ping.Serialize())
	peer.SendMulticast(msg)

	// stdin IO
	ioStruct := ui.NewIO()
	go ioStruct.StdinListen()

	// file controller
	fc := fileManager.FileController{}
	fc.Init()

	// host cache
	hc := cache.NewHostcache()

	killPeer := make(chan bool)
	ioChan := make(chan HighMessage)
	ioExpected := util.NewAtomicFlag()

	// raw peer loop
	go func() {
		for {
			select {
				case peerMsg := <-peer.RecvCh: {
					highMsg := Deserialize(peerMsg.Msg.Value)
					switch highMsg.Cmd {
						case LIST: {
							response := HighMessage{LIST_REPLY, fc.ListLocalFiles(), config.Config.Name}
							msg := network.CreateMessage(response.Serialize())
							peer.SendUnicast(msg, peerMsg.From)
						}
						case PING: {
							// update host-name cache
							highMsg := Deserialize(peerMsg.Msg.Value)
							hc.Put(highMsg.Source, peerMsg.From)
							fmt.Println("received ping!", peerMsg.From)

							// respond with a PING_REPLY so that the pinger can update his cache
							response := HighMessage{PING_REPLY, nil, config.Config.Name}
							msg := network.CreateMessage(response.Serialize())
							peer.SendUnicast(msg, peerMsg.From)
						}
						case LIST_REPLY: {
							if ioExpected.True() {
								ioChan <- highMsg
							}
						}
						case PING_REPLY: {
							hc.Put(highMsg.Source, peerMsg.From)
							fmt.Println("received ping reply!", peerMsg.From)
						}
					}
				}
				case <-killPeer: {
					return
				}
			}
		}
	}()


	// logic client loop
	// switch case based on the type of command received as input
	for {
		inp := <-ioStruct.IOChan
		switch inp.Op {
			case ui.LIST_CMD: {
				ioExpected.Set(true)

				m := HighMessage{LIST, nil, config.Config.Name}
				peer.SendMulticast(network.CreateMessage(m.Serialize()))

				timer := time.NewTimer(time.Second * WAIT_DURATION)
				replies := []HighMessage{}
				replyWait := true
				for {
					select {
						case highMsg := <-ioChan: {
							if highMsg.Cmd == LIST_REPLY {
								replies = append(replies, highMsg)
							}
						}
						case <-timer.C: {
							replyWait = false
							ioExpected.Set(false)
							break
						}
					}
					if !replyWait {
						break
					}
				}

				table := Merge(fc, replies)
				table.Display()
			}
			case ui.GET_CMD: {
			}
			case ui.LIST_LOCAL_CMD: {
				l := fc.ListLocalFiles()
				for _, f := range l {
					fmt.Println(f)
				}
			}
			case ui.LIST_USERS_CMD: {
				users := hc.Cache()
				for k, v := range users {
					fmt.Println(k, v)
				}
			}
		}

		ioStruct.IOCmdComplete <- true
	}
}

// Merge merges the lists of files obtained from the machines on the LAN that replied
// (since multiple machines may have the same files (or portions of them)
func Merge(fc fileManager.FileController, replies []HighMessage) ui.StdoutTable {
	// merge
	m := make(map[string][]*fileManager.MyFile)
	users := make(map[string][]string)
	for _, r := range replies {
		for _, f := range r.Files {
			m[f.FullHash] = append(m[f.FullHash], f)
			users[f.FullHash] = append(users[f.FullHash], r.Source)
		}
	}

	table := ui.StdoutTable{}

	for k, v := range m {
		r := ui.StdoutRecord{}
		r.FullHash = k
		mergedHashBitVector := &fileManager.BitVector{}
		var representativeFile *fileManager.MyFile
		for _, f := range v {
			r.Names = append(r.Names, f.Name)
			mergedHashBitVector.BitVectorOr(f.HashBitVector)
			representativeFile = f
		}
		r.PercentComplete = mergedHashBitVector.PercentSet(representativeFile.NumBlocks())
		localCopy := fc.FileFromHash(k)
		if localCopy == nil {
			r.PercentLocal = 0
			r.MaxComplete = r.PercentComplete
		} else {
			r.PercentLocal = localCopy.PercentComplete()
			mergedHashBitVector.BitVectorOr(localCopy.HashBitVector)
			r.MaxComplete = mergedHashBitVector.PercentSet(representativeFile.NumBlocks())
		}
		r.Users = users[k]
		table.Records = append(table.Records, r)
	}
	return table
}
