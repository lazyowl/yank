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

// IO is used for communicating with application
type IO struct {
	IOChan chan Command
	IOCmdComplete chan bool
}

// NewIO creates a new IO struct
func NewIO() *IO {
	ch := make(chan Command)
	ch1 := make(chan bool)
	return &IO{ch, ch1}
}

// StdinListen listens for stdin on the custom prompt and sends it up via IOChan.
// It waits for an ACK from IOCmdComplete before displaying the next prompt
func (i *IO) StdinListen() {
	for {
		fmt.Printf("$ ")
		bio := bufio.NewReader(os.Stdin)
		line, _, _:= bio.ReadLine()
		toks := strings.Split(string(line), " ")
		switch toks[0] {
			case "ls": {
				i.IOChan <- Command{LIST_CMD, ""}
				<-i.IOCmdComplete
			}
			case "get": {
				if len(toks) > 1 {
					i.IOChan <- Command{GET_CMD, toks[1]}
					<-i.IOCmdComplete
				}
			}
			case "lls": {
				i.IOChan <- Command{LIST_LOCAL_CMD, ""}
				<-i.IOCmdComplete
			}
			case "lu": {
				i.IOChan <- Command{LIST_USERS_CMD, ""}
				<-i.IOCmdComplete
			}
			default: {
				fmt.Println("Invalid command!")
			}
		}
	}
}
