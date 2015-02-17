package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"log"
)

var Config Configuration

type Configuration struct {
	Name string
	PublicDir string
	MetaDir string
}

// reads the configuration file and exposes a Configuration object for others to use
func ReadConfig(path string) {
	abspath, _ := filepath.Abs(path)
	file, fileErr := os.Open(abspath)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err := decoder.Decode(&Config)
	if err != nil {
		log.Fatal(fileErr)
	}
	Config.PublicDir, _ = filepath.Abs(Config.PublicDir)
	Config.MetaDir, _ = filepath.Abs(Config.MetaDir)

	os.Mkdir(Config.PublicDir, 0777)
	os.Mkdir(Config.MetaDir, 0777)
}
