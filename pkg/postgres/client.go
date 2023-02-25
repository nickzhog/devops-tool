package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
}

func NewClient(ctx context.Context, maxAttempts int, dsn string) (pool Client, err error) {
	delay := 5 * time.Second
	for maxAttempts > 0 {
		if err = func() error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			pool, err = pgxpool.Connect(ctx, dsn)
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
