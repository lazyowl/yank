package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONFIG_FILE_PATH = "lanfile_config.txt"

var Config Configuration

type Configuration struct {
	Name string
	Public_dir string
}

func Read_config() error {
	file, file_err := os.Open(CONFIG_FILE_PATH)
	if file_err != nil {
		return file_err
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err := decoder.Decode(&Config)
	if err != nil {
		return err
	}
	fmt.Println(Config.Name)
	return nil
}
