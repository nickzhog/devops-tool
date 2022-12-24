package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nickzhog/practicum-metric/internal/server/config"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClient(ctx context.Context, maxAttempts int, cfg config.Config) (pool *pgxpool.Pool, err error) {
	delay := 5 * time.Second
	for maxAttempts > 0 {
		if err = func() error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			pool, err = pgxpool.Connect(ctx, cfg.Settings.DatabaseDSN)
			return err
		}(); err != nil {
			time.Sleep(delay)
			maxAttempts--
			continue
		}
		return pool, nil
	}

	return nil, err
}
