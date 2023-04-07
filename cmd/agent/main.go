package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nickzhog/devops-tool/internal/agent/agent"
	"github.com/nickzhog/devops-tool/internal/agent/config"
	"github.com/nickzhog/devops-tool/pkg/logging"
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

	a := agent.NewAgent(cfg, logger)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		t := time.NewTicker(cfg.Settings.PollInterval)
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Trace("update metrics is stopped")
				return
			case <-t.C:
				a.UpdateMetrics()
			}
		}
	}()

	wg.Add(1)
	go func() {
		t := time.NewTicker(cfg.Settings.ReportInterval)
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Trace("send metrics is stopped")
				return
			case <-t.C:
				a.SendMetricsHTTP(ctx)

				if cfg.Settings.PortGRPC != "" {
					a.SendMetricsGRPC(ctx)
				}
			}
		}
	}()

	wg.Wait()
	logger.Trace("graceful shutdown")
}
