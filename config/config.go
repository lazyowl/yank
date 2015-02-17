package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"log"
)

const (
	PING_INTERVAL = 10		// multicast every PING_INTERVAL seconds if something has changed
	MAX_FILE_REQUESTS = 20	// maximum outstanding requests
	REQUEST_TTL = 5
	DEFAULT_PUBLIC_DIR = "publicDir"
	DEFAULT_META_DIR = "metaDir"
)

var Config Configuration

type Configuration struct {
	Name string
	PublicDir string
	MetaDir string

	PingInterval int
	MaxFileRequests int
	RequestTTL int
}

// reads the configuration file and exposes a Configuration object for others to use
func ReadConfig(path, name string, ping, maxfilereq, requestttl int) {
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

	// override read values if necessary
	if name != "" {
		log.Printf("[INFO] using name override (%s)\n", name)
		Config.Name = name
	} else if Config.Name == "" {
		log.Fatal("Name cannot be empty")
	}

	if ping > 0 {
		Config.PingInterval = ping
	} else if Config.PingInterval <= 0 {
		log.Printf("[INFO] using default ping interval (%ds)\n", PING_INTERVAL)
		Config.PingInterval = PING_INTERVAL
	}

	if maxfilereq > 0 {
		Config.MaxFileRequests = maxfilereq
	} else if Config.MaxFileRequests <= 0 {
		log.Printf("[INFO] using max file requests (%d)\n", MAX_FILE_REQUESTS)
		Config.MaxFileRequests = MAX_FILE_REQUESTS
	}

	if requestttl > 0 {
		Config.RequestTTL = requestttl
	} else if Config.RequestTTL <= 0 {
		log.Printf("[INFO] using default request ttl (%ds)\n", REQUEST_TTL)
		Config.RequestTTL = REQUEST_TTL
	}

	os.Mkdir(Config.PublicDir, 0777)
	os.Mkdir(Config.MetaDir, 0777)
}
