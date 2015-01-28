package util

import "sync"

type AtomicFlag struct {
	flag bool
	lock *sync.Mutex
}

func New_atomicflag() *AtomicFlag {
	a := AtomicFlag{}
	a.flag = false
	a.lock = &sync.Mutex{}
	return &a
}

func (f *AtomicFlag) Set(val bool) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.flag = val
}

func (f *AtomicFlag) Get() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.flag
}
