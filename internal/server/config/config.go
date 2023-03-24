package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Settings struct {
		Address string `env:"ADDRESS"`

		PostgresStorage struct {
			DatabaseDSN string `env:"DATABASE_DSN"`
		}

		RedisStorage struct {
			Addr     string `env:"REDIS_ADDR"`
			Password string `env:"REDIS_PASSWORD"`
			DB       int    `env:"REDIS_DB"`
		}

		StoreFile     string        `env:"STORE_FILE"`
		Restore       bool          `env:"RESTORE"`
		StoreInterval time.Duration `env:"STORE_INTERVAL"`

		Key string `env:"ENCRYPTION_KEY"` // ключ для вычисления хэша метрики

		CryptoKey string `env:"CRYPTO_KEY"` // путь до файла с приватным ключем (ассиметричное шифрование)
	}
}

func GetConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Settings.Address, "a", ":8080", "address for server listen")

	flag.StringVar(&cfg.Settings.PostgresStorage.DatabaseDSN, "d", "", "database dsn")

	flag.StringVar(&cfg.Settings.RedisStorage.Addr, "redis_addr", "", "redis address")
	flag.StringVar(&cfg.Settings.RedisStorage.Password, "redis_psw", "", "redis password")
	flag.IntVar(&cfg.Settings.RedisStorage.DB, "redis_db", 0, "redis database")

	flag.StringVar(&cfg.Settings.StoreFile, "f", "/tmp/devops-metrics-db.json", "file path for save and load metrics")
	flag.BoolVar(&cfg.Settings.Restore, "r", true, "restore latest values")
	flag.DurationVar(&cfg.Settings.StoreInterval, "i", time.Second, "interval for file update")

	flag.StringVar(&cfg.Settings.Key, "k", "", "key for calculate hash of metric")

	flag.StringVar(&cfg.Settings.CryptoKey, "crypto-key", "", "private.key path for RSA encryption")

	flag.Parse()

	env.Parse(&cfg.Settings)
	env.Parse(&cfg.Settings.PostgresStorage)
	env.Parse(&cfg.Settings.RedisStorage)

	return cfg
}
