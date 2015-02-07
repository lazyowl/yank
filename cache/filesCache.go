package cache

import "yank/fileManager"

type UserFileCache struct {
	cache map[string]map[string]fileManager.MyFile	// user -> (hash -> list of myfiles)
}

func NewUserFileCache() *UserFileCache {
	uf := UserFileCache{}
	uf.cache = make(map[string]map[string]fileManager.MyFile)
	return &uf
}

func (uf *UserFileCache) Put(user string, f fileManager.MyFile) {
	_, found := uf.cache[user]
	if !found {
		uf.cache[user] = make(map[string]fileManager.MyFile)
	}
	uf.cache[user][f.FullHash] = f
}

func (uf *UserFileCache) GetAll() map[string]map[string]fileManager.MyFile {
	return uf.cache
}

func (uf *UserFileCache) ClearUser(user string) {
	_, found := uf.cache[user]
	if found {
		delete(uf.cache, user)
	}
}

func (uf *UserFileCache) GetExistingByHash(hash string) map[string]fileManager.MyFile {
	userMap := make(map[string]fileManager.MyFile)
	for u, v := range uf.cache {
		f, found := v[hash]
		if found {
			userMap[u] = f
		}
	}
	return userMap
}
