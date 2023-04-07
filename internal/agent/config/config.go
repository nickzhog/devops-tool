package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		ConfigFileJSON string `env:"CONFIG" json:"config_file_json,omitempty"`

		PollInterval   time.Duration `yaml:"poll_interval" env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
		ReportInterval time.Duration `yaml:"report_interval" env:"REPORT_INTERVAL" json:"report_interval,omitempty"`
		Address        string        `yaml:"address" env:"ADDRESS" json:"address,omitempty"`
		PortGRPC       string        `yaml:"port_grpc" env:"PORT_GRPC" json:"port_grpc,omitempty"`
		Key            string        `yaml:"key" env:"KEY" json:"key,omitempty"`    // ключ для вычисления хэша метрики
		CryptoKey      string        `env:"CRYPTO_KEY" json:"crypto_key,omitempty"` // путь до файла с публичным ключем (ассиметричное шифрование)
	} `yaml:"settings"`
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Settings.ConfigFileJSON, "c", "", "path to json config file")
	env.Parse(&cfg.Settings.ConfigFileJSON)
	if cfg.Settings.ConfigFileJSON != "" {
		cfg = parseFileJSON(cfg.Settings.ConfigFileJSON)
	}

	flag.DurationVar(&cfg.Settings.PollInterval, "p", time.Second*2, "interval for update metrics")
	flag.DurationVar(&cfg.Settings.ReportInterval, "r", time.Second*10, "interval for send metrics")

	flag.StringVar(&cfg.Settings.Address, "a", "http://127.0.0.1:8080", "address for sending metrics")
	flag.StringVar(&cfg.Settings.PortGRPC, "g", "", "grpc port")

	flag.StringVar(&cfg.Settings.Key, "k", "", "key for calculate hash of metric")
	flag.StringVar(&cfg.Settings.CryptoKey, "crypto-key", "", "public.key path for RSA encryption")

	flag.Parse()

	env.Parse(&cfg.Settings)

	return cfg
}

func parseFileJSON(path string) (cfg *Config) {
	cfg = &Config{}
	file, err := os.ReadFile(path)
	if err != nil {
		return
	}

	json.Unmarshal(file, &cfg.Settings)
	return
}
