package service

import (
	"context"

	"github.com/nickzhog/devops-tool/pkg/metric"
)

type Storage interface {
	UpsertMetric(ctx context.Context, metric metric.Metric) error
	FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error)
	ExportToJSON(ctx context.Context) ([]byte, error)
	ImportFromJSON(ctx context.Context, data []byte) error
	Ping(ctx context.Context) error
}
