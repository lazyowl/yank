package cache

import "yank/fileManager"

type FileCacheValue struct {
	Names map[string][]string			// map of name -> users
	ChunkOwners map[uint][]string	// map of chunk id -> list of users
	MergedBitVector *fileManager.BitVector
}

func newFileCacheValue() FileCacheValue {
	fv := FileCacheValue{}
	fv.Names = make(map[string][]string)
	fv.ChunkOwners = make(map[uint][]string)
	temp := fileManager.BitVectorZero()
	fv.MergedBitVector = &temp
	return fv
}

type FileListCache struct {
	cache map[string]FileCacheValue
}

func NewFileListCache() *FileListCache {
	f := FileListCache{}
	f.cache = make(map[string]FileCacheValue)
	return &f
}

func (f *FileListCache) PutChunk(hash string, chunkPos uint, user string) {
	if _, found := f.cache[hash]; !found {
		f.cache[hash] = newFileCacheValue()
	}
	f.cache[hash].ChunkOwners[chunkPos] = append(f.cache[hash].ChunkOwners[chunkPos], user)
	f.cache[hash].MergedBitVector.SetBit(chunkPos)
}

func (f *FileListCache) PutName(hash string, name string, user string) {
	if _, found := f.cache[hash]; !found {
		f.cache[hash] = newFileCacheValue()
	}
	f.cache[hash].Names[name] = append(f.cache[hash].Names[name], user)
}

func (f *FileListCache) GetAll() map[string]FileCacheValue {
	return f.cache
}

func (f *FileListCache) Get(hash string) FileCacheValue {
	return f.cache[hash]
}
