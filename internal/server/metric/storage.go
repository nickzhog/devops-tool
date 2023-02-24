package metric

import "context"

type Storage interface {
	UpsertMetric(ctx context.Context, metrics *Metric) error
	FindMetric(ctx context.Context, name, mtype string) (Metric, error)
	ExportToJSON(ctx context.Context) ([]byte, error)
	ImportFromJSON(ctx context.Context, data []byte) error
}
