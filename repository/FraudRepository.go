package repository

import (
	"context"
	"fraud-detection/pkg/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const (
	insertTransactionSQL = `
INSERT INTO transactions (
  transaction_id,
  amount,
  account_id,
  location,
  transaction_time,
  elapsed_time,
  frequency,
  fraud_label
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING transaction_id;`

	getByAccountSQL = `
SELECT
  transaction_id,
  amount,
  account_id,
  location,
  transaction_time,
  elapsed_time,
  frequency,
  fraud_label
FROM transactions
WHERE account_id = $1
ORDER BY transaction_time DESC;`

	getAllSQL = `
SELECT
  transaction_id,
  amount,
  account_id,
  location,
  transaction_time,
  elapsed_time,
  frequency,
  fraud_label
FROM transactions
ORDER BY transaction_time DESC;`
)

func InsertTransaction(ctx context.Context, db DB, t transaction.Transaction) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var id string
	err := db.QueryRow(ctx, insertTransactionSQL,
		t.TransactionID,
		t.Amount,
		t.AccountID,
		t.Location,
		t.TransactionTime,
		t.ElapsedTime,
		t.Frequency,
		t.FraudLabel,
	).Scan(&id)
	return id, err
}

func GetTransactionsByAccountID(ctx context.Context, db DB, accountID string) ([]transaction.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, getByAccountSQL, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []transaction.Transaction
	for rows.Next() {
		var t transaction.Transaction
		if err := rows.Scan(
			&t.TransactionID,
			&t.Amount,
			&t.AccountID,
			&t.Location,
			&t.TransactionTime,
			&t.ElapsedTime,
			&t.Frequency,
			&t.FraudLabel,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func GetTransactions(ctx context.Context, db DB) ([]transaction.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, getAllSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []transaction.Transaction
	for rows.Next() {
		var t transaction.Transaction
		if err := rows.Scan(
			&t.TransactionID,
			&t.Amount,
			&t.AccountID,
			&t.Location,
			&t.TransactionTime,
			&t.ElapsedTime,
			&t.Frequency,
			&t.FraudLabel,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
