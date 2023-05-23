package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/jinzhu/configor"
)

func MakeConfig() (*AppConfig, error) {
	var configFilePath string

	newConfig := configor.New(&configor.Config{})

	getConfigFile := func() string {
		switch getEnvironment() {
		case "test":
			configFilePath = "./config.test.yml"
		default:
			configFilePath = "./config.yml"
		}
		_, filename, _, _ := runtime.Caller(0)
		if strings.Contains(strings.ToLower(os.Args[0]), "test") {
			return path.Join(path.Dir(filename), "./config.test.yml")
		}
		return path.Join(path.Dir(filename), configFilePath)
	}

	conf := new(AppConfig)
	err := newConfig.Load(conf, getConfigFile())
	return conf, err
}

func getEnvironment() string {
	if env := os.Getenv("CONFIGOR_ENV"); env != "" {
		return env
	}

	return "dev"
}

type Server struct {
	Port int    `yaml:"Port"`
	Host string `yaml:"Host"`
}
type Database struct {
	DbType string `yaml:"DbType"`
	Port   string `yaml:"Port"`
	Host   string `yaml:"Host"`
	DbName string `yaml:"DbName"`
}

type AppConfig struct {
	Server   Server   `yaml:"Server"`
	Database Database `yaml:"Database"`
}
