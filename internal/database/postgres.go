package database

import (
	"arena/internal/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(context.Background(), cfg.DataBaseURL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
