package config

import (
	"sync"

	"github.com/caarlos0/env"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Setting struct {
		Address string `yaml:"address" env-default:":8080" env:"ADDRESS,required"`
	} `yaml:"settings"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		cfgEnv := Config{}
		err := env.Parse(&cfgEnv.Setting)
		if err == nil {
			instance.Setting = cfgEnv.Setting
			return
		}

		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			_, _ = cleanenv.GetDescription(instance, nil)
			// log.Fatal("load config err:", err)
			instance.Setting.Address = ":8080"
		}
	})

	return instance
}
