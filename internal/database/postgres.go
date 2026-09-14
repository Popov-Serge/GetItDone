package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())

	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
