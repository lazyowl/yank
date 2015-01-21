package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const CONFIG_FILE_PATH = "./lanfile_config.txt"

var Config Configuration

type Configuration struct {
	Name string
	Public_dir string
	Meta_dir string
}

func Read_config() error {
	abspath, _ := filepath.Abs(CONFIG_FILE_PATH)
	file, file_err := os.Open(abspath)
	if file_err != nil {
		return file_err
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err := decoder.Decode(&Config)
	if err != nil {
		return err
	}
	Config.Public_dir, _ = filepath.Abs(Config.Public_dir)
	fmt.Println(Config.Name)
	fmt.Println(Config.Public_dir)
	fmt.Println(Config.Meta_dir)
	return nil
}
