package server

import (
	"context"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
)

type Server struct {
	Logger  *logging.Logger
	cfg     *config.Config
	storage service.Storage
}

func NewServer(logger *logging.Logger, cfg *config.Config, storage service.Storage) *Server {
	return &Server{
		Logger:  logger,
		cfg:     cfg,
		storage: storage,
	}
}

func (s *Server) FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error) {
	m, err := s.storage.FindMetric(ctx, name, mtype)
	if s.cfg.Settings.Key != "" {
		m.Hash = m.GetHash(s.cfg.Settings.Key)
	}
	return m, err
}

func (s *Server) UpsertMetric(ctx context.Context, m metric.Metric) error {
	if s.cfg.Settings.Key != "" && !m.IsValidHash(s.cfg.Settings.Key) {
		return metric.ErrWrongHash
	}

	return s.storage.UpsertMetric(ctx, m)
}

func (s *Server) UpsertMany(ctx context.Context, metrics []metric.Metric) error {
	if s.cfg.Settings.Key != "" {
		for _, m := range metrics {
			if !m.IsValidHash(s.cfg.Settings.Key) {
				return metric.ErrWrongHash
			}
		}
	}
	return s.storage.ImportMetrics(ctx, metrics)
}

func (s *Server) FindAll(ctx context.Context) ([]metric.Metric, error) {
	return s.storage.ExportMetrics(ctx)
}

func (s *Server) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
