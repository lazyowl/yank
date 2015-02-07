package network

import (
	"net"
	"yank/fileManager"
	"encoding/json"
)

/* Lower level message */
type Message []byte

type Response struct {
	Msg Message
	From net.Addr
}


/* Higher level message */
type DataTuple struct {
	Position int
	Data []byte
}

func NewDataTuple(pos int, dat []byte) DataTuple {
	return DataTuple{pos, dat}
}

type CmdMessage struct {
	// control messaging
	Cmd int
	Files []fileManager.MyFile
	Source string

	// data messaging
	Hash string
	RequestedChunkNumbers []int
	ReturnedDataChunks []DataTuple
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
