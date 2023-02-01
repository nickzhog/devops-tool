package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		PollInterval   time.Duration `yaml:"poll_interval" env:"POLL_INTERVAL"`
		ReportInterval time.Duration `yaml:"report_interval" env:"REPORT_INTERVAL"`
		Address        string        `yaml:"address" env:"ADDRESS"`
		Key            string        `yaml:"key" env:"KEY"`
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.DurationVar(&cfg.Settings.PollInterval, "p", time.Second*2, "interval for update metrics")
	flag.DurationVar(&cfg.Settings.ReportInterval, "r", time.Second*10, "interval for send metrics")
	flag.StringVar(&cfg.Settings.Address, "a", "http://127.0.0.1:8080", "address for sending metrics")
	flag.StringVar(&cfg.Settings.Key, "k", "", "encription key")

	flag.Parse()

	env.Parse(&cfg.Settings)

	return cfg
}
