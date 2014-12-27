package main

import (
	"fmt"
	"lanfile/network/client"
	"lanfile/network/server"
	"lanfile/network/message"
)

func main() {
	fmt.Println("Hello!")

	ch_c := make(chan message.Message)
	c, err_c := client.NewClient(ch_c)
	fmt.Println(err_c)
	go c.StartLoop()

	ch_s := make(chan message.Message)
	s, err_s := server.NewServer(ch_s)
	fmt.Println(err_s)
	go s.StartLoop()


	for {
		var s string
		fmt.Scanf("%s", &s)
		m := message.Message{s}
		ch_c <- m
	}
}
