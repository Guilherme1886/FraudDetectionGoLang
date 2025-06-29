package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func Connect() (*pgx.Conn, error) {
	var url = "postgres://fraud_user:fraud_pass@localhost:5432/fraud_db"
	return pgx.Connect(context.Background(), url)
}
