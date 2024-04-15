package config

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Mysql map[string]struct {
		Host     string
		Db       string
		User     string
		Password string
	}
	Rpc map[string]struct {
		Url      string
		User     string
		Password string
	}
	MinConfirmation map[string]int
	OrdGenesisBlock map[string]int64
	OrdProtocolName map[string]string
}

var _config = &Config{}

func Init(configPath string) {
	if configPath == "" {
		_, configFilename, _, _ := runtime.Caller(0)
		configPath = path.Join(path.Dir(configFilename), "config.toml")
	}
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("read config file %s error:%s", configPath, err.Error())
		os.Exit(1)
	} else {
		if err2 := toml.Unmarshal(bytes, _config); err2 != nil {
			log.Printf("toml unmarshal %s error:%s", string(bytes), err2.Error())
			os.Exit(1)
		}
	}
}

func Instance() *Config {
	return _config
}
