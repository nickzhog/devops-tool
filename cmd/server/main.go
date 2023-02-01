package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nickzhog/devops-tool/internal/server/compress"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/handlers"
	"github.com/nickzhog/devops-tool/internal/server/metric/cache"
	"github.com/nickzhog/devops-tool/internal/server/metric/db"
	"github.com/nickzhog/devops-tool/internal/server/postgres"
	"github.com/nickzhog/devops-tool/internal/server/storagefile"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger()
	logger.Tracef("config: %+v", cfg)

	handlerData := &handlers.Handler{
		Logger: logger,
		Cfg:    cfg,
	}

	if cfg.Settings.DatabaseDSN != "" {
		var err error
		handlerData.ClientDB, err = postgres.NewClient(context.Background(), 2, cfg)
		if err != nil {
			logger.Tracef("db error: %s", err.Error())
		}
		handlerData.Storage = db.NewRepository(handlerData.ClientDB, logger, cfg)
	} else {
		handlerData.Storage = cache.NewMemStorage()
	}

	if cfg.Settings.StoreFile != "" {
		storagefile.NewStorageFile(cfg, logger, handlerData.Storage)
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(compress.GzipMiddleWare)

	r.Get("/", handlerData.IndexHandler)
	r.Get("/ping", handlerData.PingHandler)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlerData.SelectFromBody)
		r.Get("/{metric_type}/{name}", handlerData.SelectFromURL)
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlerData.UpdateFromBody)
		r.Post("/{metric_type}/{name}/{value}", handlerData.UpdateFromURL)
	})

	// batch update
	r.Post("/updates/", handlerData.UpdateMany)

	logger.Fatal(http.ListenAndServe(cfg.Settings.Address, r))
}
