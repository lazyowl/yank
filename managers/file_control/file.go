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
	Hash_bit_vector uint64	// bit vector describing how many of these are present on the current machine
	Size int
}

func (f *MyFile) String() string {
	return fmt.Sprintf("name=%s, %%complete=%d%%, size=%d", f.Name, f.Percent_complete(), f.Size)
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

func (f *MyFile) Percent_complete() int {
	count := 0
	one := uint64(0x01)
	var num_blocks int
	if f.Size % CHUNK_SIZE == 0 {
		num_blocks = f.Size / CHUNK_SIZE
	} else {
		num_blocks = f.Size / CHUNK_SIZE + 1
	}
	for i := 0; i < num_blocks; i++ {
		if (f.Hash_bit_vector & one) > 0 {
			count++
		}
		one <<= 1
	}
	return 100 * count / num_blocks
}
