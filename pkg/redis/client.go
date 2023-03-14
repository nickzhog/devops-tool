package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, addr, pwd string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}

	return rdb
}
