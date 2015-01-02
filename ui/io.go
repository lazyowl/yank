package ui

import (
	"fmt"
)

type IO struct {
	IO_chan chan string
}

func NewIO() *IO {
	ch := make(chan string)
	return &IO{ch}
}

func (i *IO) StdinListen() {
	for {
		var s string
		fmt.Scanf("%s", &s)
		i.IO_chan <- s
	}
}
