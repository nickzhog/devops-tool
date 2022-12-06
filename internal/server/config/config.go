package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Setting struct {
		Port int `yaml:"poll_interval" env-default:"8080"`
	} `yaml:"settings"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			_, _ = cleanenv.GetDescription(instance, nil)
			// log.Fatal("load config err:", err)
			instance.Setting.Port = 8080
		}
	})
	return instance
}
