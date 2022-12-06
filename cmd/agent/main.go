package main

import (
	"sync"
	"time"

	"github.com/nickzhog/practicum-metric/internal/agent/config"
	"github.com/nickzhog/practicum-metric/internal/agent/metric"
)

func main() {
	cfg := config.GetConfig()

	var metrics metric.Metrics
	metrics.InitMetrics()
	metrics.UpdateMetrics()
	metrics.SendMetrics(cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			metrics.UpdateMetrics()
			time.Sleep(time.Millisecond * time.Duration(cfg.Intervals.PollInterval))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			metrics.SendMetrics(cfg)
			time.Sleep(time.Millisecond * time.Duration(cfg.Intervals.ReportInterval))
		}
	}()

	wg.Wait()
}
