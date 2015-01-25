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
		if toks[0] == "ls" {
			i.IO_chan <- Command{LIST_CMD, ""}
			<-i.IO_cmd_complete
		} else if toks[0] == "get" && len(toks) > 1 {
			i.IO_chan <- Command{GET_CMD, toks[1]}
			<-i.IO_cmd_complete
		} else if toks[0] == "lls" {
			i.IO_chan <- Command{LIST_LOCAL_CMD, ""}
			<-i.IO_cmd_complete
		} else {
			fmt.Println("Invalid command!")
		}
	}
}
