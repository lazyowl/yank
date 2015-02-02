// Package hostcache contains logic for a caching mechanism for storing names -> IP address pairs
package cache

import (
	"net"
)

type Hostcache struct {
	cache map[string]net.Addr
}

// NewHostcache returns a new host cache
func NewHostcache() *Hostcache {
	h := Hostcache{}
	h.cache = make(map[string]net.Addr)
	return &h
}

// Put
func (h *Hostcache) Put(host string, ip net.Addr) {
	h.cache[host] = ip
}

// Get
func (h *Hostcache) Get(host string) net.Addr {
	return h.cache[host]
}

// Cache returns the entire contents of the cache
func (h *Hostcache) Cache() map[string]net.Addr {
	return h.cache
}
