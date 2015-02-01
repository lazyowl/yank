package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const CONFIG_FILE_PATH = "./config.txt"

var Config Configuration

type Configuration struct {
	Name string
	PublicDir string
	MetaDir string
}

// ReadConfig reads the configuration file and exposes a Configuration object for others to use
func ReadConfig() error {
	abspath, _ := filepath.Abs(CONFIG_FILE_PATH)
	file, fileErr := os.Open(abspath)
	if fileErr != nil {
		return fileErr
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err := decoder.Decode(&Config)
	if err != nil {
		return err
	}
	Config.PublicDir, _ = filepath.Abs(Config.PublicDir)
	Config.MetaDir, _ = filepath.Abs(Config.MetaDir)
	return nil
}
