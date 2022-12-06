package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nickzhog/practicum-metric/internal/server/config"
	"github.com/nickzhog/practicum-metric/internal/server/metric"
)

func main() {
	cfg := config.GetConfig()
	storage := metric.NewMemStorage()

	r := mux.NewRouter()

	tpl, err := template.ParseGlob("pages/*.html")
	if err != nil {
		log.Printf("pages err: %v", err.Error())
	}

	handlerData := &metric.Handler{
		Data: storage,
		Tpl:  tpl,
	}

	r.HandleFunc("/", handlerData.IndexHandler)

	r.HandleFunc("/update/{metric_type}/{name}/{value}", handlerData.UpdateHandler)
	r.HandleFunc("/value/{metric_type}/{name}", handlerData.SelectHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", cfg.Setting.Port), r))
}
