package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nickzhog/practicum-metric/internal/server/config"
	"github.com/nickzhog/practicum-metric/internal/server/metric"
)

func main() {
	cfg := config.GetConfig()
	storage := metric.NewMemStorage()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
		})
	})

	tpl, err := template.ParseGlob("pages/*.html")
	if err != nil {
		log.Printf("pages err: %v", err.Error())
	}

	handlerData := &metric.Handler{
		Data: storage,
		Tpl:  tpl,
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

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", cfg.Setting.Port), r))
}
