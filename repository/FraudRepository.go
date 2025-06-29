package repository

import (
	"context"
	"fraud-detection/pkg/transaction"
	"github.com/jackc/pgx/v5"
)

func InsertTransaction(ctx context.Context, conn *pgx.Conn, t transaction.Transaction) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO transactions (account_id, amount, timestamp, location, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`, t.AccountID, t.Amount, t.Timestamp, t.Location, t.IPAddress)

	return err
}

func GetTransactionsByAccountID(ctx context.Context, conn *pgx.Conn, accountID string) ([]transaction.Transaction, error) {
	rows, err := conn.Query(ctx, `
		SELECT id, account_id, amount, timestamp, location, ip_address
		FROM transactions
		WHERE account_id = $1
-- 		ORDER BY timestamp DESC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []transaction.Transaction
	for rows.Next() {
		var t transaction.Transaction
		err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Timestamp, &t.Location, &t.IPAddress)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func GetTransactions(ctx context.Context, conn *pgx.Conn) ([]transaction.Transaction, error) {
	rows, err := conn.Query(ctx, `
		SELECT id, account_id, amount, timestamp, location, ip_address
		FROM transactions
-- 		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []transaction.Transaction
	for rows.Next() {
		var t transaction.Transaction
		err := rows.Scan(&t.ID, &t.AccountID, &t.Amount, &t.Timestamp, &t.Location, &t.IPAddress)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
