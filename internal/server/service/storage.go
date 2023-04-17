package service

import (
	"context"

	"github.com/nickzhog/devops-tool/pkg/metric"
)

type Storage interface {
	UpsertMetric(ctx context.Context, metric metric.Metric) error
	FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error)
	ExportMetrics(ctx context.Context) ([]metric.Metric, error)
	ImportMetrics(ctx context.Context, metrics []metric.Metric) error
	Ping(ctx context.Context) error
}
