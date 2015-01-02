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
)

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
	l := fc.List_public_files()
	fmt.Println(l)

	// main loop
	for {
		select {
			case server_msg := <-s.Recv_ch: {
				fmt.Println(server_msg)
			}
			case client_msg := <-c.Recv_ch: {
				fmt.Println("client listen:", client_msg)
			}
			case inp := <-io_struct.IO_chan: {
				c.SendMulticast(message.CreateQueryMessage(inp))
			}
		}
	}
}
