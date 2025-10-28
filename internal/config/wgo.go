package config

import (
	"os"
	"sync"
)

// import "github.com/zhufuyi/sponge/pkg/conf"

var (
	config     *Config
	onceConfig sync.Once
)

// func Init(configFile string, fs ...func()) error {
// config = &Config{}
// return conf.Parse(configFile, config, fs...)
// }

func Get() *Config {
	if config == nil {
		onceConfig.Do(func() {
			config = &Config{
				LoggerFile: os.Getenv("HOME") + "/.local/share/wgo/log/wgo.log",
			}
		})
	}
	return config
}

type Config struct {
	LoggerFile string `yaml:"logger_file" json:"logger_file"`
}
