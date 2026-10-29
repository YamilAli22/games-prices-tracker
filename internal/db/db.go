package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	DATABASE_URL := os.Getenv("DATABASE_URL")
	conn, err := pgxpool.New(
		context.Background(),
		DATABASE_URL)

	if err != nil {
		return nil, err
	}
	
	return conn, nil
}
