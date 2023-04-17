package web

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/server"
	"github.com/nickzhog/devops-tool/internal/server/web/middleware"
	"github.com/nickzhog/devops-tool/pkg/encryption"
)

func Serve(ctx context.Context, srv server.Server, cfg *config.Config) {
	handlerData := NewHandler(srv)

	r := chi.NewRouter()

	// r.Use(chimiddleware.Logger)
	if cfg.Settings.TrustedSubnet != "" {
		r.Use(chimiddleware.RealIP)

		_, ipNet, err := net.ParseCIDR(cfg.Settings.TrustedSubnet)
		if err != nil {
			srv.Logger.Fatal(err)
		}
		r.Use(middleware.CheckIP(ipNet, srv.Logger))
	}

	r.Use(middleware.GzipCompress)
	r.Use(middleware.GzipDecompress)

	if cfg.Settings.CryptoKey != "" {
		key, err := encryption.NewPrivateKey(cfg.Settings.CryptoKey)
		if err != nil {
			srv.Logger.Fatal(err)
		}
		r.Use(middleware.RequestDecryptMiddleWare(key, srv.Logger))
	}

	r.Mount("/debug", chimiddleware.Profiler())

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

	httpSrv := &http.Server{
		Addr:    cfg.Settings.Address,
		Handler: r,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%+s\n", err)
		}
	}()

	srv.Logger.Tracef("server started")

	<-ctx.Done()

	srv.Logger.Tracef("server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctxShutDown); err != nil {
		srv.Logger.Fatalf("server Shutdown Failed:%+s", err)
	}

	srv.Logger.Tracef("server exited properly")
}
