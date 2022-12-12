package config

import (
	"sync"
	"time"

	"github.com/caarlos0/env"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Setting struct {
		Address       string        `yaml:"address" env-default:":8080" env:"ADDRESS,required"`
		StoreInterval time.Duration `yaml:"store_interval" env-default:"300s" env:"STORE_INTERVAL"`
		StoreFile     string        `yaml:"store_file" env-default:"/tmp/devops-metrics-db.json" env:"STORE_FILE"`
		Restore       bool          `yaml:"restore" env-default:"true" env:"RESTORE"`
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

			instance.Setting.StoreInterval = time.Second * 300
			instance.Setting.StoreFile = "/tmp/devops-metrics-db.json"
			instance.Setting.Restore = true
		}
	})

	return instance
}
