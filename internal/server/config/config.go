package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		Address       string        `yaml:"address" env:"ADDRESS"`
		DatabaseDSN   string        `yaml:"database_dsn" env:"DATABASE_DSN"`
		StoreFile     string        `yaml:"store_file" env:"STORE_FILE"`
		Restore       bool          `yaml:"restore" env:"RESTORE"`
		StoreInterval time.Duration `yaml:"store_interval" env:"STORE_INTERVAL"`
		Key           string        `yaml:"key" env:"KEY"`
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Settings.Address, "a", ":8080", "address for server listen")
	flag.StringVar(&cfg.Settings.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.Settings.StoreFile, "f", "/tmp/devops-metrics-db.json", "file for save and load metrics")
	flag.BoolVar(&cfg.Settings.Restore, "r", true, "restore latest values")
	flag.DurationVar(&cfg.Settings.StoreInterval, "i", time.Second, "interval for file update")
	flag.StringVar(&cfg.Settings.Key, "k", "", "encription key")
	flag.Parse()

	env.Parse(&cfg.Settings)

	return cfg
}
