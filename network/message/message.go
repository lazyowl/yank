package message

import (
	"encoding/json"
	"net"
)

const (
	QUERY = iota
	FETCH
)

type Message struct {
	Type int
	Value string
}

type Response struct {
	Msg Message
	From net.Addr
}

func ToJson(m Message) []byte {
	b, _ := json.Marshal(m)
	return b
}

func FromJson(b []byte) Message {
	var msg Message
	json.Unmarshal(b, &msg)
	return msg
}

func CreateQueryMessage(query string) Message {
	m := Message{QUERY, query}
	return m
}

func CreateMessage(typ int, val string) Message {
	m := Message{typ, val}
	return m
}
