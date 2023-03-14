package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/internal/server/metric/cache"
	"github.com/nickzhog/devops-tool/internal/server/metric/db"
	"github.com/nickzhog/devops-tool/internal/server/metric/redis"
	"github.com/nickzhog/devops-tool/internal/server/migration"
	"github.com/nickzhog/devops-tool/internal/server/storagefile"
	"github.com/nickzhog/devops-tool/internal/server/web"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/postgres"
	redis_client "github.com/nickzhog/devops-tool/pkg/redis"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger()
	logger.Tracef("config: %+v", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		oscall := <-c
		logger.Tracef("system call:%+v", oscall)
		cancel()
	}()

	var storage metric.Storage

	switch {

	case cfg.Settings.PostgresStorage.DatabaseDSN != "":
		logger.Trace("postgres storage")
		err := migration.Migrate(cfg.Settings.PostgresStorage.DatabaseDSN)
		if err != nil {
			logger.Fatalf("migration error: %s", err.Error())
		}
		postgresClient, err := postgres.NewClient(ctx, 2, cfg.Settings.PostgresStorage.DatabaseDSN)
		if err != nil {
			logger.Fatalf("db error: %s", err.Error())
		}
		storage = db.NewRepository(postgresClient, logger, cfg)

	case cfg.Settings.RedisStorage.Addr != "":
		logger.Trace("redis storage")
		redisClient := redis_client.NewClient(ctx,
			cfg.Settings.RedisStorage.Addr,
			cfg.Settings.RedisStorage.Password,
			cfg.Settings.RedisStorage.DB)
		storage = redis.NewRepository(redisClient, logger, cfg)

	default:
		logger.Trace("inmemory storage")
		storage = cache.NewMemStorage()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		srv := web.PrepareServer(logger, cfg, storage)
		if err := web.Serve(ctx, logger, srv); err != nil {
			logger.Errorf("failed to serve: %s", err.Error())
		}
		wg.Done()
	}()

	if cfg.Settings.StoreFile != "" {
		wg.Add(1)
		go func() {
			storagefile.NewStorageFile(ctx, cfg, logger, storage).StartUpdate(ctx)
			wg.Done()
		}()
	}
	wg.Wait()
}
