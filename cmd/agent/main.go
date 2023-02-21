package main

import (
	"context"
	"os"
	"os/signal"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		oscall := <-c
		logger.Tracef("system call:%+v", oscall)
		cancel()
	}()

	agent := metric.NewAgent()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		t := time.NewTicker(cfg.Settings.PollInterval)
		select {
		case <-ctx.Done():
			logger.Trace("update metrics is stopped")
		case <-t.C:
			agent.UpdateMetrics()
		}
		wg.Done()
	}()

	go func() {
		t := time.NewTicker(cfg.Settings.ReportInterval)
		select {
		case <-ctx.Done():
			logger.Trace("send metrics is stopped")
		case <-t.C:
			agent.SendMetrics(cfg, logger)
		}
		wg.Done()
	}()

	wg.Wait()
}
