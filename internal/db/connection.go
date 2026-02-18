package db

import (
	"context"
	"os"
	"fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type DBServer struct {
	DB *pgxpool.Pool
}

func InitDB(ctx context.Context) (*DBServer, error) {
	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("Нет подключения к БД: %w", err)
	}

	ServerDB := &DBServer{
		DB: pool,
	}

	return ServerDB, nil
}