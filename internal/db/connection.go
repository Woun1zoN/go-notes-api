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
	user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    dbname := os.Getenv("DB_NAME")

	conn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=require",
		user, password, host, port, dbname,
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