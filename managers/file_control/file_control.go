package file_control

import (
	"os"
	"lanfile/managers/config"
	"gopkg.in/fsnotify.v1"
	"log"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type File_controller struct {
	watcher *fsnotify.Watcher
}


func (fc File_controller) Init() chan bool {
	var err error
	fc.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)
	go func() {
		log.Println("started watcher!")
		for {
			select {
				case event := <-fc.watcher.Events: {
					log.Println(event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						fc.Generate_MyFile_entry(event.Name)
					}
					if event.Op&fsnotify.Remove == fsnotify.Remove {
						fc.Destroy_file(event.Name)
					}
				}
				case err := <-fc.watcher.Errors: {
					log.Println(err)
				}
				case <-done: {
					break
				}
			}
		}
	}()
	err = fc.watcher.Add(config.Config.Public_dir)
	fmt.Println("watcher listening to ", config.Config.Public_dir)
	if err != nil {
		log.Fatal(err)
	}
	return done
}

func (fc File_controller) get_MyFile_from_name(name string) (*MyFile, error) {
	file_contents, err := ioutil.ReadFile(filepath.Join(config.Config.Meta_dir, name))
	if err != nil {
		// this might happen when a file is simply created but not written to
		fmt.Println("err!", err)
		return nil, err
	}
	return Deserialize(file_contents), nil
}

func (fc File_controller) write_MyFile(f *MyFile) error {
	fmt.Println("writing to file:", f.Name)
	return ioutil.WriteFile(filepath.Join(config.Config.Meta_dir, f.Name), f.Serialize(), 0666)
}


// assumes the entire file is present locally (to be used when locally creating a new public file)
func (fc File_controller) Generate_MyFile_entry(filename string) (*MyFile, error) {
	fmt.Println("Generate_MyFile_entry:", filename)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("readfile")
		log.Fatal(err)
	}
	full_hash := Hash(b)

	// bit vector has all ones since we have the file
	file := MyFile{filepath.Base(filename), full_hash, 0xffffffff, len(b)}
	fmt.Println("FILENAME IS ", file.Name, file.Size)

	err = fc.write_MyFile(&file)
	if err != nil {
		log.Fatal(err)
	}

	return &file, nil
}

func (fc File_controller) Destroy_file(name string) bool {
	file_contents, err := ioutil.ReadFile(name)
	if err != nil {
		// this might happen when a file is simply created but not written to
		fmt.Println("err!", err)
		return false
	}
	f := Deserialize(file_contents)
	f.Hash_bit_vector = 0x00
	err = fc.write_MyFile(f)
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func (fc File_controller) List_local_files() []*MyFile {
	dir, err := os.Open(config.Config.Public_dir)
	if err != nil {
		log.Fatal(err)
	}
	names, _ := dir.Readdirnames(0)
	files := []*MyFile{}
	for _, k := range names {
		f, err := fc.get_MyFile_from_name(k)
		if err == nil {
			files = append(files, f)
		}
	}
	return files
}
