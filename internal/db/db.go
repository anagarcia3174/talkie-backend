package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error creating db pool: %w", err)
	}
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = dbpool.Ping(ctxTimeout)
	if err != nil {
		dbpool.Close()
		return nil, fmt.Errorf("error pinging db pool: %w", err)
	}

	return dbpool, nil
}
