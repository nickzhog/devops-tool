package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		DatabaseDSN   string        `yaml:"database_dsn" env:"DATABASE_DSN"`
		Address       string        `yaml:"address" env:"ADDRESS"`
		StoreInterval time.Duration `yaml:"store_interval" env:"STORE_INTERVAL"`
		StoreFile     string        `yaml:"store_file" env:"STORE_FILE"`
		Restore       bool          `yaml:"restore" env:"RESTORE"`
		Key           string        `yaml:"key" env:"KEY"`
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Settings.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&cfg.Settings.Address, "a", ":8080", "address for server listen")
	flag.BoolVar(&cfg.Settings.Restore, "r", true, "restore latest values")
	flag.StringVar(&cfg.Settings.StoreFile, "f", "/tmp/devops-metrics-db.json", "file for db")
	flag.DurationVar(&cfg.Settings.StoreInterval, "i", time.Second, "interval for db update")
	flag.StringVar(&cfg.Settings.Key, "k", "", "encription key")
	flag.Parse()

	env.Parse(&cfg.Settings)

	return cfg
}
