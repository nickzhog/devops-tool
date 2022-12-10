package config

import (
	"sync"
	"time"

	"github.com/caarlos0/env"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Settings struct {
		PollInterval   time.Duration `yaml:"poll_interval" env-default:"2s" env:"POLL_INTERVAL,required"`
		ReportInterval time.Duration `yaml:"report_interval" env-default:"10s" env:"REPORT_INTERVAL,required"`
		Address        string        `yaml:"address" env-default:"http://127.0.0.1:8080" env:"ADDRESS,required"`
	} `yaml:"settings"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}

		cfgEnv := Config{}
		err := env.Parse(&cfgEnv.Settings)
		if err == nil {
			instance.Settings = cfgEnv.Settings
			return
		}

		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			_, _ = cleanenv.GetDescription(instance, nil)
			// log.Fatal("load config err:", err)
			instance.Settings.PollInterval = time.Second * 2
			instance.Settings.ReportInterval = time.Second * 10
			instance.Settings.Address = "http://127.0.0.1:8080"
		}
	})
	return instance
}
