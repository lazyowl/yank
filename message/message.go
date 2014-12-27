package message

import (
	"encoding/json"
)

type Message struct {
	Value string
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
