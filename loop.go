package main

import (
	"os"
	"fmt"
	"lanfile/network/client"
	"lanfile/network/server"
	"lanfile/network/message"
	"lanfile/managers/config"
	"lanfile/managers/file_control"
	"lanfile/ui"
	"encoding/json"
)



type HighMessage struct {
	Cmd int
	Files []*file_control.MyFile
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
	fmt.Println(err_read)

	// client
	ch_c := make(chan message.Response)
	c, err_c := client.NewClient(ch_c)
	if err_c != nil {
		fmt.Println(err_c)
		os.Exit(1)
	}
	go c.ListenUnicast()
	go c.ListenMulticast()

	// server
	ch_s := make(chan message.Response)
	s, err_s := server.NewServer(ch_s)
	if err_s != nil {
		fmt.Println(err_s)
		os.Exit(1)
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
				fmt.Println(high_msg)
				if high_msg.Cmd == 0 {
					response := HighMessage{1, fc.List_local_files()}
					msg := message.Message{0, response.Serialize()}
					s.SendUnicast(server_msg.From, msg)
				}
			}
			case client_msg := <-c.Recv_ch: {
				fmt.Println("client listen:", client_msg)
				high_msg := Deserialize(client_msg.Msg.Value)
				fmt.Println(high_msg)
				if high_msg.Cmd == 1 {
					for _, f := range high_msg.Files {
						fmt.Println(f)
					}
				}
			}
			case inp := <-io_struct.IO_chan: {
				if inp == "ls" {
					files := fc.List_local_files()
					fmt.Println(len(files), "files")
					for _, f := range files {
						fmt.Println(f)
					}
					fmt.Println("===")
				}
				m := HighMessage{}
				c.SendMulticast(message.CreateMessage(0, m.Serialize()))
			}
		}
	}
}
