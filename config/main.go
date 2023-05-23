package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port" envconfig:"SERVER_PORT" default:"4444"`
		Addr string `yaml:"addr" envconfig:"SERVER_HOST" default:"0.0.0.0"`
	} `yaml:"server"`
	Database struct {
		URI  string `yaml:"uri" envconfig:"DB_URI" required:"true"`
		Name string `yaml:"name" envconfig:"DB_NAME" required:"true"`
	} `yaml:"database"`
}

var (
	once sync.Once
	Cfg  Config
)

func Init() {
	once.Do(func() {
		readFile(&Cfg)
		readEnv(&Cfg)
	})
}

func processError(err error) {
	fmt.Println(err)
}

func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
		return
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}
