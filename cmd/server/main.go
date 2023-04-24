package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/server"
	"github.com/nickzhog/devops-tool/internal/server/server/grpc"
	web "github.com/nickzhog/devops-tool/internal/server/server/http"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/internal/server/service/cache"
	"github.com/nickzhog/devops-tool/internal/server/service/db"
	"github.com/nickzhog/devops-tool/internal/server/service/redis"
	"github.com/nickzhog/devops-tool/internal/server/storagefile"
	"github.com/nickzhog/devops-tool/migration"
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
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		oscall := <-c
		logger.Tracef("system call:%+v", oscall)
		cancel()
	}()

	var storage service.Storage
	switch {

	case cfg.PostgresStorage.DatabaseDSN != "":
		logger.Trace("postgres storage")
		err := migration.Migrate(cfg.PostgresStorage.DatabaseDSN)
		if err != nil {
			logger.Fatalf("migration error: %s", err.Error())
		}
		postgresClient, err := postgres.NewClient(ctx, 2, cfg.PostgresStorage.DatabaseDSN)
		if err != nil {
			logger.Fatalf("db error: %s", err.Error())
		}
		storage = db.NewRepository(postgresClient, logger, cfg)

	case cfg.RedisStorage.Addr != "":
		logger.Trace("redis storage")
		redisClient := redis_client.NewClient(ctx,
			cfg.RedisStorage.Addr,
			cfg.RedisStorage.Password,
			cfg.RedisStorage.DB)
		storage = redis.NewRepository(redisClient, logger, cfg)

	default:
		logger.Trace("inmemory storage")
		storage = cache.NewMemStorage()
	}

	srv := server.NewServer(logger, cfg, storage)

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		web.Serve(ctx, *srv, cfg)
		wg.Done()
	}()

	go func() {
		grpc.Serve(ctx, *srv, cfg)
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
