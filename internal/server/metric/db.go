package metric

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nickzhog/devops-tool/internal/server/postgresql"
	"github.com/nickzhog/devops-tool/pkg/logging"
)

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(client postgresql.Client, logger *logging.Logger) Storage {
	q := `
	CREATE TABLE IF NOT EXISTS public.metrics (
		id text not null, 
		type text not null,
		value double precision,
		delta BIGINT
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

func (r *repository) FindMetric(ctx context.Context, name, mtype string) (Metric, bool) {
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
	m := Metric{ID: name, MType: mtype}
	err := r.client.QueryRow(ctx, q, mtype, name).Scan(
		&delta, &value)

	if err != nil {
		r.logger.Errorf("metric find err:%s", err.Error())
		return Metric{}, false
	}

	switch mtype {
	case CounterType:
		if !delta.Valid {
			return Metric{}, false
		}
		m.Delta = &delta.Int64
	case GaugeType:
		if !value.Valid {
			return Metric{}, false
		}
		m.Value = &value.Float64
	}

	return m, true
}

func (r *repository) UpsertMetric(ctx context.Context, metric *Metric) (err error) {
	if metric.MType == CounterType {
		oldMetric, exist := r.FindMetric(ctx, metric.ID, metric.MType)
		if exist {
			delta := *metric.Delta + *oldMetric.Delta
			metric.Delta = &delta
		}
	} else if metric.MType != GaugeType {
		return errors.New("wrong metric type")
	}

	q := `
	UPDATE metrics 
	SET
		id=$1, type=$2, value=$3, delta=$4  
	WHERE 
		id=$1 and type=$2;
	`
	commandTag, err := r.client.Exec(ctx, q,
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

func (r *repository) ImportFromJSON(ctx context.Context, data []byte) error {
	var metrics []Metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}

	// todo: transaction
	for _, v := range metrics {
		err = r.UpsertMetric(ctx, &v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repository) ExportToJSON(ctx context.Context) ([]byte, error) {
	q := `
		SELECT
		id, type, delta, value 
		FROM public.metrics;
	`

	var metrics []Metric
	rows, err := r.client.Query(ctx, q)
	if err != nil {
		r.logger.Errorf("metrics find err:%s", err.Error())
		return []byte(``), err
	}

	for rows.Next() {
		var m Metric

		var delta sql.NullInt64
		var value sql.NullFloat64

		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			r.logger.Errorf("metrics parse:%s", err.Error())
			return []byte(``), err
		}

		switch m.MType {
		case CounterType:
			if !delta.Valid {
				return []byte(``), fmt.Errorf("null delta for %s", m.ID)
			}
			m.Delta = &delta.Int64
		case GaugeType:
			if !value.Valid {
				return []byte(``), fmt.Errorf("null value for %s", m.ID)
			}
			m.Value = &value.Float64
		}

		metrics = append(metrics, m)
	}

	return json.Marshal(metrics)
}
