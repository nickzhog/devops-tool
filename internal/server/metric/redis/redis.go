package redis

import (
	"context"
	"encoding/json"

	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/redis/go-redis/v9"
)

type repository struct {
	client *redis.Client
	logger *logging.Logger
	cfg    *config.Config
}

func NewRepository(client *redis.Client, logger *logging.Logger, cfg *config.Config) *repository {
	return &repository{
		client: client,
		logger: logger,
		cfg:    cfg,
	}
}

func (r *repository) Ping(ctx context.Context) error {
	return nil
}

func (r *repository) FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error) {
	result, err := r.client.Get(ctx, prepareKey(name, mtype)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return metric.Metric{}, metric.ErrNoResult
		}
		return metric.Metric{}, err
	}

	var answer metric.Metric
	err = json.Unmarshal(result, &answer)
	if err != nil {
		return metric.Metric{}, err
	}

	return answer, nil
}

func (r *repository) UpsertMetric(ctx context.Context, m metric.Metric) error {
	if m.MType == metric.CounterType {
		mcurrent, err := r.FindMetric(ctx, m.ID, metric.CounterType)
		if err != nil && err != metric.ErrNoResult {
			return err
		}
		*m.Delta += *mcurrent.Delta
	}
	return r.client.Set(ctx, prepareKey(m.ID, m.MType), m.Marshal(), 0).Err()
}

func (r *repository) ImportFromJSON(ctx context.Context, data []byte) error {
	var metrics []metric.Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	for _, m := range metrics {
		if r.cfg.Settings.Key != "" {
			if !m.IsValidHash(r.cfg.Settings.Key) {
				return metric.ErrWrongHash
			}
		}

		err = r.UpsertMetric(ctx, m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) ExportToJSON(ctx context.Context) ([]byte, error) {
	keys, err := r.client.Keys(ctx, "metric:*").Result()
	if err != nil {
		return nil, err
	}

	metrics := make([]metric.Metric, 0)
	for _, k := range keys {
		data, err := r.client.Get(ctx, k).Bytes()
		if err != nil {
			return nil, err
		}
		var m metric.Metric
		err = json.Unmarshal(data, &m)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}

	return data, nil
}
