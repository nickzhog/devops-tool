package redis

import (
	"context"
	"log"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Settings.RedisStorage.Addr,
		Password: cfg.Settings.RedisStorage.Password,
		DB:       cfg.Settings.RedisStorage.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	return rdb
}
