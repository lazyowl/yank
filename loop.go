package main

import (
	"fmt"
	"lanfile/network/client"
	"lanfile/network/server"
	"lanfile/network/message"
	"lanfile/managers/config"
	"lanfile/managers/file_control"
	"lanfile/ui"
	"encoding/json"
	"log"
)

const (
	LIST = 0
	LIST_REPLY = 1
)

type HighMessage struct {
	Cmd int
	Files []*file_control.MyFile
	Source string
}

func (m HighMessage) Serialize() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func Deserialize(b string) HighMessage {
	var msg HighMessage
	json.Unmarshal([]byte(b), &msg)
	return msg
}

func main() {
	fmt.Println("Hello!")

	// read configuration
	err_read := config.Read_config()
	if err_read != nil {
		log.Fatal(err_read)
	}

	// client
	ch_c := make(chan message.Response)
	c, err_c := client.NewClient(ch_c)
	if err_c != nil {
		log.Fatal(err_c)
	}
	go c.ListenUnicast()
	go c.ListenMulticast()

	// server
	ch_s := make(chan message.Response)
	s, err_s := server.NewServer(ch_s)
	if err_s != nil {
		log.Fatal(err_s)
	}
	go s.Listen()

	// stdin IO
	io_struct := ui.NewIO()
	go io_struct.StdinListen()

	fc := file_control.File_controller{}
	fc.Init()

	// main loop
	for {
		select {
			case server_msg := <-s.Recv_ch: {
				high_msg := Deserialize(server_msg.Msg.Value)
				if high_msg.Cmd == LIST {
					response := HighMessage{LIST_REPLY, fc.List_local_files(), config.Config.Name}
					msg := message.Message{0, response.Serialize()}
					s.SendUnicast(server_msg.From, msg)
				}
			}
			case client_msg := <-c.Recv_ch: {
				high_msg := Deserialize(client_msg.Msg.Value)
				if high_msg.Cmd == LIST_REPLY {
					fmt.Println("===>", high_msg.Source)
					for _, f := range high_msg.Files {
						fmt.Println(f)
					}
				}
			}
			case inp := <-io_struct.IO_chan: {
				if inp == "ls" {
					m := HighMessage{LIST, nil, config.Config.Name}
					c.SendMulticast(message.CreateMessage(0, m.Serialize()))
				}
			}
		}
	}
}
