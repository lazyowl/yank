package fileManager

import (
	"os"
	"yank/config"
	"gopkg.in/fsnotify.v1"
	"log"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// Type FileController provides functions for interacting with the Public and Meta dirs
type FileController struct {
	watcher *fsnotify.Watcher
}

// Init initializes the file controller and sets up the watcher
func NewFileController() *FileController {
	fc := FileController{}
	var err error
	fc.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan bool)
	go func() {
		for {
			select {
				case event := <-fc.watcher.Events: {
					log.Println(event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						fc.generateMyFileEntry(event.Name)
					}
					if event.Op&fsnotify.Remove == fsnotify.Remove {
						fc.DestroyFile(event.Name)
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
	err = fc.watcher.Add(config.Config.PublicDir)
	fmt.Println("watcher listening to ", config.Config.PublicDir)
	if err != nil {
		log.Fatal(err)
	}
	return &fc
}

// getMyFileFromName returns a MyFile pointer from the filename
func (fc FileController) getMyFileFromName(name string) (*MyFile, error) {
	fileContents, err := ioutil.ReadFile(filepath.Join(config.Config.MetaDir, name))
	if err != nil {
		// this might happen when a file is simply created but not written to
		fmt.Println("err!", err)
		return nil, err
	}
	f := Deserialize(fileContents)
	f.lock = &sync.Mutex{}
	return f, nil
}

// writeMyFile writes the MyFile into the Meta directory
func (fc *FileController) writeMyFile(f *MyFile) error {
	fmt.Println("writing to file:", f.Name)
	return ioutil.WriteFile(filepath.Join(config.Config.MetaDir, f.Name), f.Serialize(), 0666)
}


// generateMyFileEntry assumes the entire file is present locally (to be used when locally creating a new public file)
func (fc *FileController) generateMyFileEntry(filename string) (*MyFile, error) {
	fmt.Println("GenerateMyFileEntry:", filename)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("readfile")
		log.Fatal(err)
	}
	fullHash := Hash(b)

	// bit vector has all ones since we have the file
	file := NewMyFile()
	file.Name = filepath.Base(filename)
	file.FullHash = fullHash
	file.HashBitVector = BitVectorOnes()
	file.Size = len(b)
	fmt.Println("FILENAME IS ", file.Name, file.Size, file.HashBitVector)

	err = fc.writeMyFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return file, nil
}


// DestroyFile destroys the file
func (fc *FileController) DestroyFile(name string) bool {
	fileContents, err := ioutil.ReadFile(name)
	if err != nil {
		// this might happen when a file is simply created but not written to
		fmt.Println("err!", err)
		return false
	}
	f := Deserialize(fileContents)
	f.HashBitVector = BitVectorZero()
	err = fc.writeMyFile(f)
	if err != nil {
		log.Fatal(err)
	}
	return true
}

// ListLocalFiles returns a list of MyFile pointers corresponding to files present in local Public dir
func (fc *FileController) ListLocalFiles() []MyFile {
	dir, err := os.Open(config.Config.PublicDir)
	if err != nil {
		log.Fatal(err)
	}
	names, _ := dir.Readdirnames(0)
	files := []MyFile{}
	for _, k := range names {
		f, err := fc.getMyFileFromName(k)
		if err == nil {
			files = append(files, *f)
		}
	}
	return files
}

// FileFromHash returns MyFile pointer from a full hash (TODO improve this)
func (fc *FileController) FileFromHash(hash string) *MyFile {
	files := fc.ListLocalFiles()
	for _, f := range files {
		if f.FullHash == hash {
			return &f
		}
	}
	return nil
}

func (fc *FileController) CreateEmptyFile(name string, hash string, size int) (*MyFile, error) {
	file, err := os.Create(filepath.Join(config.Config.PublicDir, name))
	if err != nil {
		return nil, err
	}
	truncErr := file.Truncate(int64(size))
	if truncErr != nil {
		return nil, truncErr
	}
	f, err := fc.generateMyFileEntry(filepath.Join(config.Config.PublicDir, name))
	if err != nil {
		return nil, err
	}
	f.HashBitVector = BitVectorZero()	// change to an all zeros bit vector
	file.Close()
	return f, nil
}
