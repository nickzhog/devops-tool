package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	PostgresStorage struct {
		DatabaseDSN string `env:"DATABASE_DSN"`
	}

	RedisStorage struct {
		Addr     string `env:"REDIS_ADDR"`
		Password string `env:"REDIS_PASSWORD"`
		DB       int    `env:"REDIS_DB"`
	}

	Settings struct {
		Address     string `env:"ADDRESS"`
		AddressGRPC string `env:"ADDRESS_GRPC" json:"ADDRESS_GRPC,omitempty"`

		StoreFile     string        `env:"STORE_FILE"`
		Restore       bool          `env:"RESTORE"`
		StoreInterval time.Duration `env:"STORE_INTERVAL"`

		TrustedSubnet string `env:"TRUSTED_SUBNET"`

		Key string `env:"ENCRYPTION_KEY"` // ключ для вычисления хэша метрики

		CryptoKey string `env:"CRYPTO_KEY"` // путь до файла с приватным ключем (ассиметричное шифрование)

	}
}

func GetConfig() *Config {
	cfg := new(Config)
	flag.StringVar(&cfg.Settings.AddressGRPC, "g", ":3200", "grpc port")
	flag.StringVar(&cfg.Settings.Address, "a", ":8080", "address for server listen")

	flag.StringVar(&cfg.PostgresStorage.DatabaseDSN, "d", "", "database dsn")

	flag.StringVar(&cfg.RedisStorage.Addr, "redis_addr", "", "redis address")
	flag.StringVar(&cfg.RedisStorage.Password, "redis_psw", "", "redis password")
	flag.IntVar(&cfg.RedisStorage.DB, "redis_db", 0, "redis database")

	flag.StringVar(&cfg.Settings.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path for save and load metrics")
	flag.BoolVar(&cfg.Settings.Restore, "r", true, "restore latest values")
	flag.DurationVar(&cfg.Settings.StoreInterval, "i", time.Second, "interval for file update")

	flag.StringVar(&cfg.Settings.TrustedSubnet, "t", "", "trusted subnet")

	flag.StringVar(&cfg.Settings.Key, "k", "", "key for calculate hash of metric")

	flag.StringVar(&cfg.Settings.CryptoKey, "crypto-key", "", "private.key path for RSA encryption")

	flag.Parse()

	env.Parse(&cfg.Settings)
	env.Parse(&cfg.PostgresStorage)
	env.Parse(&cfg.RedisStorage)

	return cfg
}
