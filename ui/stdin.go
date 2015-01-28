package ui

import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

const (
	LIST_LOCAL_CMD = -1
	LIST_CMD = 0
	GET_CMD = 1
	LIST_USERS_CMD = 2
)

type Command struct {
	Op int
	Arg string
}

type IO struct {
	IO_chan chan Command
	IO_cmd_complete chan bool
}

func NewIO() *IO {
	ch := make(chan Command)
	ch1 := make(chan bool)
	return &IO{ch, ch1}
}

func (i *IO) StdinListen() {
	for {
		fmt.Printf("$ ")
		bio := bufio.NewReader(os.Stdin)
		line, _, _:= bio.ReadLine()
		toks := strings.Split(string(line), " ")
		switch toks[0] {
			case "ls": {
				i.IO_chan <- Command{LIST_CMD, ""}
				<-i.IO_cmd_complete
			}
			case "get": {
				if len(toks) > 1 {
					i.IO_chan <- Command{GET_CMD, toks[1]}
					<-i.IO_cmd_complete
				}
			}
			case "lls": {
				i.IO_chan <- Command{LIST_LOCAL_CMD, ""}
				<-i.IO_cmd_complete
			}
			case "lu": {
				i.IO_chan <- Command{LIST_USERS_CMD, ""}
				<-i.IO_cmd_complete
			}
			default: {
				fmt.Println("Invalid command!")
			}
		}
	}
}
