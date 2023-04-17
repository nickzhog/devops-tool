package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/service"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/metric"
	"github.com/nickzhog/devops-tool/pkg/postgres"
)

var _ service.Storage = (*repository)(nil)

type repository struct {
	client postgres.Client
	logger *logging.Logger
	cfg    *config.Config
}

func NewRepository(client postgres.Client, logger *logging.Logger, cfg *config.Config) *repository {
	return &repository{
		client: client,
		logger: logger,
		cfg:    cfg,
	}
}

func (r *repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)
}

func (r *repository) FindMetric(ctx context.Context, name, mtype string) (metric.Metric, error) {
	q := `
		SELECT
		 	delta, value 
		FROM 
			public.metrics 
		WHERE 
			type = $1 and id = $2;
	`

	var delta sql.NullInt64
	var value sql.NullFloat64
	m := metric.Metric{ID: name, MType: mtype}
	err := r.client.QueryRow(ctx, q, mtype, name).Scan(
		&delta, &value)

	if err != nil {
		r.logger.Errorf("metric find err:%s", err.Error())
		if err == pgx.ErrNoRows {
			return metric.Metric{}, metric.ErrNoResult
		}

		return metric.Metric{}, err
	}

	switch mtype {
	case metric.CounterType:
		if !delta.Valid {
			return metric.Metric{}, metric.ErrNoResult
		}
		m.Delta = &delta.Int64
	case metric.GaugeType:
		if !value.Valid {
			return metric.Metric{}, metric.ErrNoResult
		}
		m.Value = &value.Float64
	}

	return m, nil
}

func (r *repository) UpsertMetric(ctx context.Context, metric metric.Metric) (err error) {
	q := `
	INSERT 
	INTO metrics
		(id, type, value, delta) 
	VALUES 
		($1, $2, $3, $4)
	ON CONFLICT (id,type) DO UPDATE 
	SET value=$3, delta=metrics.delta+$4;
	`
	_, err = r.client.Exec(ctx, q,
		metric.ID, metric.MType, metric.Value, metric.Delta)

	if err != nil {
		r.logger.Trace(err)
	}

	return
}

func (r *repository) ImportMetrics(ctx context.Context, metrics []metric.Metric) error {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := `
	INSERT 
	INTO metrics
		(id, type, value, delta) 
	VALUES 
		($1, $2, $3, $4)
	ON CONFLICT (id,type) DO UPDATE 
	SET value=$3, delta=metrics.delta+$4;
	`

	batch := &pgx.Batch{}
	for _, v := range metrics {
		batch.Queue(q, v.ID, v.MType, v.Value, v.Delta)
	}

	result := tx.Conn().SendBatch(ctx, batch)
	err = result.Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) ExportMetrics(ctx context.Context) ([]metric.Metric, error) {
	q := `
		SELECT
		id, type, delta, value 
		FROM public.metrics;
	`

	var metrics []metric.Metric
	rows, err := r.client.Query(ctx, q)
	if err != nil {
		r.logger.Errorf("metrics find err:%s", err.Error())
		return nil, err
	}

	for rows.Next() {
		var m metric.Metric

		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			r.logger.Errorf("metrics parse:%s", err.Error())
			return nil, err
		}

		switch m.MType {
		case metric.CounterType:
			if !delta.Valid {
				return nil, fmt.Errorf("null delta for %s", m.ID)
			}
			m.Delta = &delta.Int64
		case metric.GaugeType:
			if !value.Valid {
				return nil, fmt.Errorf("null value for %s", m.ID)
			}
			m.Value = &value.Float64
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}
