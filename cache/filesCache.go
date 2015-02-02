package cache

type FileCacheValue struct {
	chunkOwners map[int][]string
}

type FileListCache struct {
	cache map[string]FileCacheValue
}

func NewFileListCache() *FileListCache {
	f := FileListCache{}
	return &f
}
