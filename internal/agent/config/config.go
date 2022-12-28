package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Settings struct {
		PollInterval   time.Duration `yaml:"poll_interval" env-default:"2s" env:"POLL_INTERVAL,required"`
		ReportInterval time.Duration `yaml:"report_interval" env-default:"10s" env:"REPORT_INTERVAL,required"`
		Address        string        `yaml:"address" env-default:"http://127.0.0.1:8080" env:"ADDRESS,required"`
		Key            string        `yaml:"key" env-default:"" env:"KEY"`
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.DurationVar(&cfg.Settings.ReportInterval, "r", time.Second*10, "interval for send metrics")
	flag.DurationVar(&cfg.Settings.PollInterval, "p", time.Second*2, "interval for update metrics")
	flag.StringVar(&cfg.Settings.Address, "a", "http://127.0.0.1:8080", "address for sending metrics")
	flag.StringVar(&cfg.Settings.Key, "k", "", "encription key")

	flag.Parse()

	addr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Settings.Address = addr
	}
	reportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		dur, _ := time.ParseDuration(reportInterval)
		cfg.Settings.ReportInterval = dur
	}

	pollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		dur, _ := time.ParseDuration(pollInterval)
		cfg.Settings.PollInterval = dur
	}
	key, ok := os.LookupEnv("KEY")
	if ok {
		cfg.Settings.Key = key
	}

	return cfg
}
