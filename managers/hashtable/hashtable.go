package hashtable

import (
	"github.com/fzzy/radix/redis"
	"fmt"
)

type Hashtable struct {
	client *redis.Client
}

func New_hashtable() *Hashtable {
	h := Hashtable{}
	h.client, _ = redis.Dial("tcp", "localhost:6379")
	return &h
}

func (h *Hashtable) Put(key string, val []byte) error {
	reply := h.client.Cmd("SET", key, string(val))
	return reply.Err
}

func (h *Hashtable) Get(key string) ([]byte, error) {
	reply := h.client.Cmd("GET", key)
	if reply.Type == redis.NilReply {
		return []byte{}, fmt.Errorf("not found!")
	}
	str, err := reply.Str()
	return []byte(str), err
}

func (h *Hashtable) Delete(key string) error {
	reply := h.client.Cmd("DEL", key)
	return reply.Err
}

func (h *Hashtable) Close() error {
	return h.client.Close()
}
