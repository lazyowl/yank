package file_control

import (
	"os"
	"encoding/json"
	"lanfile/managers/hashtable"
	"lanfile/managers/config"
)

type File_controller struct {
}

type MyFile struct {
	Name string				// file name
	Full_hash string		// hash of entire file
	Hashes []string			// list of file chunk hashes (in order)
	Hash_bit_vector uint64	// bit vector describing how many of these are present on the current machine
}

func (f MyFile) Serialize() []byte {
	byteSlice, _ := json.Marshal(f)
	return byteSlice
}

func Deserialize(serial []byte) MyFile {
	var f MyFile
	json.Unmarshal(serial, &f)
	return f
}

func get_MyFile_from_name(name string) (MyFile, error) {
	f := MyFile{}
	return f, nil
}

func (f MyFile) generate_bit_vector() {
	db := hashtable.Hashtable{}
	var vec uint64 // default initialization to 0
	for i, k := range f.Hashes {
		_, err := db.Get(k)
		if err == nil {
			vec |= (1 << uint(i))
		}
	}
	f.Hash_bit_vector = vec
}

func (f MyFile) Percent_complete() int {
	db := hashtable.Hashtable{}
	count := 0
	for _, k := range f.Hashes {
		_, err := db.Get(k)
		if err == nil {
			count++
		}
	}
	return count / len(f.Hashes)
}

func (fc File_controller) List_public_files() []MyFile {
	dir, _ := os.Open(config.Config.Public_dir)
	names, _ := dir.Readdirnames(0)
	files := []MyFile{}
	for _, k := range names {
		f, err := get_MyFile_from_name(k)
		if err == nil {
			files = append(files, f)
		}
	}
	return files
}
