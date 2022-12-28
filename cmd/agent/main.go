package main

import (
	"sync"
	"time"

	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/internal/agent/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.GetLogger()
	logger.Tracef("config: %+v", cfg)

	agent := metric.NewAgent()

	agent.UpdateMetrics()
	agent.SendMetrics(cfg, logger)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(cfg.Settings.PollInterval)
		for range t.C {
			agent.UpdateMetrics()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(cfg.Settings.ReportInterval)
		for range t.C {
			agent.SendMetrics(cfg, logger)
		}
	}()

	wg.Wait()
}
