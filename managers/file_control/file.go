package file_control

import (
	"encoding/json"
	"fmt"
)


/* Name should be enough to uniquely identify the file locally. */
/* Full_hash should be enough to uniquely identify the file globally */
type MyFile struct {
	Name string				// file name
	Full_hash string		// hash of entire file
	Hash_bit_vector Bit_vector	// bit vector describing how many of these are present on the current machine
	Size int
}

func (f *MyFile) String() string {
	return fmt.Sprintf("%s (%s): %dB, %%complete=%d%%", f.Name, f.Full_hash, f.Size, f.Percent_complete())
}

func (f *MyFile) Serialize() []byte {
	byteSlice, _ := json.Marshal(f)
	return byteSlice
}

func Deserialize(serial []byte) *MyFile {
	var f MyFile
	json.Unmarshal(serial, &f)
	return &f
}

func (f *MyFile) Get_num_blocks() int {
	var num_blocks int
	if f.Size % CHUNK_SIZE == 0 {
		num_blocks = f.Size / CHUNK_SIZE
	} else {
		num_blocks = f.Size / CHUNK_SIZE + 1
	}
	return num_blocks
}

func (f *MyFile) Percent_complete() int {
	return f.Hash_bit_vector.Percent_set(f.Get_num_blocks())
}
