// Package hostcache contains logic for a caching mechanism for storing names -> IP address pairs
package cache

import (
	"net"
	"sync"
)

type HostCache struct {
	cache map[string]net.Addr
	lock *sync.Mutex
}

// NewHostCache returns a new host cache
func NewHostCache() *HostCache {
	h := HostCache{}
	h.cache = make(map[string]net.Addr)
	h.lock = &sync.Mutex{}
	return &h
}

// Put
func (h *HostCache) Put(host string, ip net.Addr) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.cache[host] = ip
}

// Get
func (h *HostCache) Get(host string) net.Addr {
	return h.cache[host]
}

// Cache returns the entire contents of the cache
func (h *HostCache) Cache() map[string]net.Addr {
	return h.cache
}
