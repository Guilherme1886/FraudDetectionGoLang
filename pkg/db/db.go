package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	dbURL := "postgres://fraud_user:fraud_pass@localhost:5432/fraud_db"
	return pgxpool.New(context.Background(), dbURL)
}
