package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Settings struct {
		DatabaseDSN   string        `yaml:"database_dsn" env-default:"" env:"DATABASE_DSN"`
		Address       string        `yaml:"address" env-default:":8080" env:"ADDRESS,required"`
		StoreInterval time.Duration `yaml:"store_interval" env-default:"1s" env:"STORE_INTERVAL,required"`
		StoreFile     string        `yaml:"store_file" env-default:"/tmp/devops-metrics-db.json" env:"STORE_FILE,required"`
		Restore       bool          `yaml:"restore" env-default:"true" env:"RESTORE,required"`
		Key           string        `yaml:"key" env-default:"" env:"KEY"`
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

	// cfgEnv := Config{}
	// err := env.Parse(&cfgEnv.Settings)
	// if err == nil {
	// 	cfg.Settings = cfgEnv.Settings
	// 	return
	// }

	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		cfg.Settings.DatabaseDSN = dsn
	}

	addressEnv, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Settings.Address = addressEnv
	}
	storeIntervalEnv, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		if dur, err := time.ParseDuration(storeIntervalEnv); err == nil {
			cfg.Settings.StoreInterval = dur
		}
	}

	restoreEnv, ok := os.LookupEnv("RESTORE")
	if ok {
		switch restoreEnv {
		case "false":
			cfg.Settings.Restore = false
		case "true":
			cfg.Settings.Restore = true
		}
	}
	storeFileEnv, ok := os.LookupEnv("STORE_FILE")
	if ok {
		cfg.Settings.StoreFile = storeFileEnv
	}
	key, ok := os.LookupEnv("KEY")
	if ok {
		cfg.Settings.Key = key
	}

	// if err := cleanenv.ReadConfig("config.yml", cfg); err != nil {
	// 	_, _ = cleanenv.GetDescription(cfg, nil)
	// 		log.Fatal("load config err:", err)
	// }

	return cfg
}
