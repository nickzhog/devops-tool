package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Intervals struct {
		PollInterval   int `yaml:"poll_interval" env-default:"2000"`
		ReportInterval int `yaml:"report_interval" env-default:"10000"`
	} `yaml:"intervals"`
	SendTo struct {
		Address string `yaml:"address" env-default:"http://127.0.0.1:8080"`
	} `yaml:"send_to"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			_, _ = cleanenv.GetDescription(instance, nil)
			// log.Fatal("load config err:", err)
			instance.Intervals.ReportInterval = 10000
			instance.Intervals.PollInterval = 2000
			instance.SendTo.Address = "http://127.0.0.1:8080"
		}
	})
	return instance
}
