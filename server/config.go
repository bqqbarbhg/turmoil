package server

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
)

type ConfigStruct struct {
	Server struct {
		Port int
		Url  string
	}
	Database struct {
		Sqlite *struct {
			Path string
		}
	}
	Folders struct {
		Data string
	}
}

var Config ConfigStruct

func checkFolder(name, path string) {
	if path == "" {
		log.Fatalf("Folder '%s' not specified", name)
	}
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			abs, abserr := filepath.Abs(path)
			if abserr != nil {
				abs = path
			}
			log.Fatalf("Folder '%s': '%s' does not exist", name, abs)
		} else {
			log.Fatalf("Folder '%s': %v", name, err)
		}
	}
	if !stat.IsDir() {
		log.Fatal("Folder '%s' is not a directory")
	}
}

func loadConfig() {
	_, err := toml.DecodeFile("config.toml", &Config)
	if err != nil {
		log.Fatalf("Failed to load 'config.toml': %v", err)
	}

	if !(Config.Server.Port >= 1 && Config.Server.Port <= 65535) {
		log.Fatalf("Server.Port needs to be between 1 and 65535")
	}

	if Config.Server.Url == "" {
		log.Fatalf("Server.Url is required")
	}

	if !(Config.Database.Sqlite != nil) {
		log.Fatalf("No database driver is provided")
	}

	checkFolder("Data", Config.Folders.Data)
}
