package web

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/encryption"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func PrepareServer(logger *logging.Logger, cfg *config.Config, storage service.Storage) *http.Server {

	handlerData := NewHandlerData(logger, cfg, storage)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(gzipCompress)
	r.Use(gzipDecompress)

	if cfg.Settings.CryptoKey != "" {
		key, err := encryption.NewPrivateKey(cfg.Settings.CryptoKey)
		if err != nil {
			logger.Fatal(err)
		}
		r.Use(requestDecryptMiddleWare(key, logger))
	}

	r.Mount("/debug", middleware.Profiler())

	r.Get("/ping", handlerData.PingHandler)

	r.Get("/", handlerData.IndexHandler)

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

	return &http.Server{
		Addr:    cfg.Settings.Address,
		Handler: r,
	}
}

func Serve(ctx context.Context, logger *logging.Logger, srv *http.Server) (err error) {
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	logger.Tracef("server started")

	<-ctx.Done()

	logger.Tracef("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		logger.Fatalf("server Shutdown Failed:%+s", err)
	}

	logger.Tracef("server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	return
}
