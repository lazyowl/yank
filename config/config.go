package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONFIG_FILE_PATH = "lanfile_config.txt"

type Configuration struct {
	Name string
}

func Read_config() error {
	file, file_err := os.Open(CONFIG_FILE_PATH)
	if file_err != nil {
		return file_err
	}
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		return err
	}
	fmt.Println(configuration.Name)
	return nil
}
