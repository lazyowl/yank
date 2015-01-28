package hostcache

import (
	//"fmt"
	//"lanfile/network/messaging"
	"net"
)

type Hostcache struct {
	cache map[string]net.Addr
}

func New_hostcache() *Hostcache {
	h := Hostcache{}
	h.cache = make(map[string]net.Addr)
	return &h
}

func (h *Hostcache) Put(host string, ip net.Addr) {
	h.cache[host] = ip
}

func (h *Hostcache) Get(host string) net.Addr {
	return h.cache[host]
}

func (h *Hostcache) Get_cache() map[string]net.Addr {
	return h.cache
}
