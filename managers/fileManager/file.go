package fileManager

import (
	"encoding/json"
	"fmt"
)


// MyFile represents a file
// Name should be enough to uniquely identify the file locally.
// Full_hash should be enough to uniquely identify the file globally
type MyFile struct {
	Name string					// file name
	FullHash string				// hash of entire file
	HashBitVector BitVector		// bit vector describing how many of these are present on the current machine
	Size int					// size
}

// String returns a string representation
func (f *MyFile) String() string {
	return fmt.Sprintf("%s (%s): %dB, %%complete=%d%%", f.Name, f.FullHash, f.Size, f.PercentComplete())
}

// Serialize
func (f *MyFile) Serialize() []byte {
	byteSlice, _ := json.Marshal(f)
	return byteSlice
}

// Deserialize
func Deserialize(serial []byte) *MyFile {
	var f MyFile
	json.Unmarshal(serial, &f)
	return &f
}

// NumBlocks returns the number of chunks created (based on a fixed size chunking scheme)
func (f *MyFile) NumBlocks() int {
	var numBlocks int
	if f.Size % CHUNK_SIZE == 0 {
		numBlocks = f.Size / CHUNK_SIZE
	} else {
		numBlocks = f.Size / CHUNK_SIZE + 1
	}
	return numBlocks
}

// PercentComplete returns the percentage of the file that is available
func (f *MyFile) PercentComplete() int {
	return f.HashBitVector.PercentSet(f.NumBlocks())
}
