package metric

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nickzhog/devops-tool/internal/server/postgresql"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

type Repository interface {
	CreateTable(ctx context.Context)
	Upsert(ctx context.Context, metric MetricExport) (err error)
	FindByName(ctx context.Context, mtype, name string) (MetricExport, error)
	FindAll(ctx context.Context) ([]MetricExport, error)
}

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(client postgresql.Client, logger *logging.Logger) Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) CreateTable(ctx context.Context) {
	q := `
		CREATE TABLE IF NOT EXISTS public.metrics (
			id text not null, 
			type text not null,
			value double precision,
			delta BIGINT
		);
	`
	_, err := r.client.Exec(ctx, q)
	if err != nil {
		r.logger.Error(err)
		return
	}
}

func (r *repository) FindAll(ctx context.Context) ([]MetricExport, error) {
	q := `
		SELECT
		id, type, delta, value 
		FROM public.metrics;
	`

	var metrics []MetricExport
	rows, err := r.client.Query(ctx, q)
	if err != nil {
		r.logger.Errorf("metrics find err:%s", err.Error())
		return []MetricExport{}, err
	}

	for rows.Next() {
		var m MetricExport

		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			r.logger.Errorf("metrics parse:%s", err.Error())
			return []MetricExport{}, err
		}

		switch m.MType {
		case CounterType:
			if !delta.Valid {
				return []MetricExport{}, fmt.Errorf("null delta for %s", m.ID)
			}
			m.Delta = &delta.Int64
		case GaugeType:
			if !value.Valid {
				return []MetricExport{}, fmt.Errorf("null value for %s", m.ID)
			}
			m.Value = &value.Float64
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (r *repository) FindByName(ctx context.Context, mtype, name string) (MetricExport, error) {
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
	var m MetricExport
	err := r.client.QueryRow(ctx, q, mtype, name).Scan(
		&delta, &value)

	if err != nil {
		r.logger.Errorf("metric find err:%s", err.Error())
		return MetricExport{}, err
	}

	switch mtype {
	case CounterType:
		if !delta.Valid {
			return MetricExport{}, fmt.Errorf("null delta for %s", name)
		}
		m.Delta = &delta.Int64
	case GaugeType:
		if !value.Valid {
			return MetricExport{}, fmt.Errorf("null value for %s", name)
		}
		m.Value = &value.Float64
	}

	return m, nil
}

func (r *repository) Upsert(ctx context.Context, metric MetricExport) (err error) {
	q := `
	UPDATE metrics 
	SET
		id=$1, type=$2, value=$3, delta=$4  
	WHERE 
		id=$1 and type=$2;
	`
	commandTag, _ := r.client.Exec(ctx, q,
		metric.ID, metric.MType, metric.Value, metric.Delta)

	if err != nil {
		r.logger.Trace(err)
	}

	if commandTag.RowsAffected() != 1 {
		q = `
		INSERT INTO metrics (id, type, value, delta)
		SELECT $1, $2, $3, $4 
		WHERE NOT EXISTS (SELECT 1 FROM metrics WHERE id=$1 and type=$2);
		`
		_, err = r.client.Exec(ctx, q,
			metric.ID, metric.MType, metric.Value, metric.Delta)
	}

	return
}
