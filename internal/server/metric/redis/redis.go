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

func (r *repository) UpsertMetric(ctx context.Context, metric *metric.Metric) error {
	return r.client.Set(ctx, prepareKeyForMetric(*metric), metric.Marshal(), 0).Err()
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

		err = r.UpsertMetric(ctx, &m)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) ExportToJSON(ctx context.Context) ([]byte, error) {
	iter := r.client.Scan(ctx, 0, "prefix:*", 0).Iterator()

	var metrics []metric.Metric

	for iter.Next(ctx) {
		var m metric.Metric
		err := json.Unmarshal([]byte(iter.Val()), &m)
		if err != nil {
			return []byte(``), err
		}
	}
	if err := iter.Err(); err != nil {
		return []byte(``), err
	}

	return json.Marshal(metrics)
}
