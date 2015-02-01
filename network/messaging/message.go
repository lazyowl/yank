package messaging

import (
	"encoding/json"
	"net"
)

type Message struct {
	Value []byte
}

type Response struct {
	Msg Message
	From net.Addr
}

func Serialize(m Message) []byte {
	b, _ := json.Marshal(m)
	return b
}

func Deserialize(b []byte) Message {
	var msg Message
	json.Unmarshal(b, &msg)
	return msg
}

func CreateMessage(val []byte) Message {
	m := Message{val}
	return m
}
