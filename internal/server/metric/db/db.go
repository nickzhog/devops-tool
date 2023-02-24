package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/nickzhog/devops-tool/internal/server/config"
	"github.com/nickzhog/devops-tool/internal/server/metric"
	"github.com/nickzhog/devops-tool/pkg/logging"
	"github.com/nickzhog/devops-tool/pkg/postgres"
)

type repository struct {
	client postgres.Client
	logger *logging.Logger
	cfg    *config.Config
}

func NewRepository(client postgres.Client, logger *logging.Logger, cfg *config.Config) *repository {
	q := `
	CREATE TABLE IF NOT EXISTS public.metrics (
		id text not null, 
		type text not null,
		value double precision,
		delta BIGINT,
		PRIMARY KEY (id, type)
	);`
	_, err := client.Exec(context.TODO(), q)
	if err != nil {
		logger.Fatal(err)
	}

	return &repository{
		client: client,
		logger: logger,
	}
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

func (r *repository) UpsertMetric(ctx context.Context, metric *metric.Metric) (err error) {
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

func (r *repository) ImportFromJSON(ctx context.Context, data []byte) error {
	var metrics []metric.Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}
	if r.cfg.Settings.Key != "" {
		for _, m := range metrics {
			if !m.IsValidHash(r.cfg.Settings.Key) {
				return fmt.Errorf("not valid hash for metric: %+v", m)
			}
		}
	}

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

func (r *repository) ExportToJSON(ctx context.Context) ([]byte, error) {
	q := `
		SELECT
		id, type, delta, value 
		FROM public.metrics;
	`

	var metrics []metric.Metric
	rows, err := r.client.Query(ctx, q)
	if err != nil {
		r.logger.Errorf("metrics find err:%s", err.Error())
		return []byte(``), err
	}

	for rows.Next() {
		var m metric.Metric

		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			r.logger.Errorf("metrics parse:%s", err.Error())
			return []byte(``), err
		}

		switch m.MType {
		case metric.CounterType:
			if !delta.Valid {
				return []byte(``), fmt.Errorf("null delta for %s", m.ID)
			}
			m.Delta = &delta.Int64
		case metric.GaugeType:
			if !value.Valid {
				return []byte(``), fmt.Errorf("null value for %s", m.ID)
			}
			m.Value = &value.Float64
		}

		metrics = append(metrics, m)
	}

	return json.Marshal(metrics)
}
