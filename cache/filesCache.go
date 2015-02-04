package cache

import "yank/fileManager"

type UserFileCache struct {
	cache map[string][]fileManager.MyFile
}

func NewUserFileCache() *UserFileCache {
	uf := UserFileCache{}
	uf.cache = make(map[string][]fileManager.MyFile)
	return &uf
}

func (uf *UserFileCache) Put(user string, f fileManager.MyFile) {
	_, found := uf.cache[user]
	if !found {
		uf.cache[user] = []fileManager.MyFile{}
	}
	uf.cache[user] = append(uf.cache[user], f)
}

func (uf *UserFileCache) GetAll() map[string][]fileManager.MyFile {
	return uf.cache
}

func (uf *UserFileCache) ClearUser(user string) {
	_, found := uf.cache[user]
	if found {
		delete(uf.cache, user)
	}
}
