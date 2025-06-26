package frauddetector

import "fraud-detection/pkg/transaction"

type TransactionChecker interface {
	IsAboveLimit(transaction transaction.Transaction) bool
	OccurMultipleTransactions(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool
}

type DefaultTransactionChecker struct{}

var Checker = &DefaultTransactionChecker{}
