package config

import (
	"flag"
	"sync"
	"time"

	"github.com/caarlos0/env"
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
		address := flag.String("a", "http://127.0.0.1:8080", "address for sending metrics")
		flag.DurationVar(&instance.Settings.ReportInterval, "r", time.Second*10, "interval for send metrics")
		flag.DurationVar(&instance.Settings.PollInterval, "p", time.Second*2, "interval for update metrics")
		flag.Parse()

		instance.Settings.Address = *address

		cfgEnv := Config{}
		err := env.Parse(&cfgEnv.Settings)
		if err == nil {
			instance.Settings = cfgEnv.Settings
			return
		}

		// if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
		// 	_, _ = cleanenv.GetDescription(instance, nil)
		// 	// log.Fatal("load config err:", err)
		// 	instance.Settings.PollInterval = time.Second * 2
		// 	instance.Settings.ReportInterval = time.Second * 10
		// 	instance.Settings.Address = "http://127.0.0.1:8080"
		// }
	})
	return instance
}
