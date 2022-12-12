package main

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nickzhog/practicum-metric/internal/server/config"
	"github.com/nickzhog/practicum-metric/internal/server/metric"
	"github.com/nickzhog/practicum-metric/pkg/logging"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger()
	logger.Traceln("config:", cfg)

	storage := metric.NewMemStorage()
	db.Connect(storage, cfg, logger)

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tpl, err := template.ParseGlob("pages/*.html")
	if err != nil {
		logger.Errorf("cant load pages: %s", err.Error())
	}

	handlerData := &metric.Handler{
		Data:   storage,
		Tpl:    tpl,
		Logger: logger,
	}

	r.Get("/", handlerData.IndexHandler)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlerData.SelectFromBody)
		r.Get("/{metric_type}/{name}", handlerData.SelectHandler)
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlerData.UpdateFromBody)
		r.Post("/{metric_type}/{name}/{value}", handlerData.UpdateHandler)
	})

	logger.Fatal(http.ListenAndServe(cfg.Setting.Address, r))
}
