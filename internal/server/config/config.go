package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		Address       string        `yaml:"address" env-default:":8080" env:"ADDRESS,required"`
		StoreInterval time.Duration `yaml:"store_interval" env-default:"300s" env:"STORE_INTERVAL,required"`
		StoreFile     string        `yaml:"store_file" env-default:"/tmp/devops-metrics-db.json" env:"STORE_FILE,required"`
		Restore       bool          `yaml:"restore" env-default:"true" env:"RESTORE,required"`
		Key           string        `yaml:"key" env-default:"" env:"KEY"`
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Settings.Address, "a", ":8080", "address for server listen")
	flag.BoolVar(&cfg.Settings.Restore, "r", true, "restore latest values")
	flag.StringVar(&cfg.Settings.StoreFile, "f", "/tmp/devops-metrics-db.json", "file for db")
	flag.StringVar(&cfg.Settings.Key, "k", "", "encription key")
	flag.DurationVar(&cfg.Settings.StoreInterval, "i", time.Second*300, "interval for db update")
	flag.Parse()

	cfgEnv := Config{}
	err := env.Parse(&cfgEnv.Settings)
	if err == nil {
		cfg.Settings = cfgEnv.Settings
	}

	// if err := cleanenv.ReadConfig("config.yml", cfg); err != nil {
	// 	_, _ = cleanenv.GetDescription(cfg, nil)
	// 		log.Fatal("load config err:", err)
	// }

	return cfg
}
