package frauddetector

import (
	"fraud-detection/pkg/transaction"
	"time"
)

const MaxAmount = 1000000

func IsSuspiciousTransaction(
	transaction transaction.Transaction,
	transactions []transaction.Transaction,
	lastLocation string,
	current string,
	orderOfQuery string,
	checker TransactionChecker,
) bool {
	return checker.IsAboveLimit(transaction) ||
		checker.OccurMultipleTransactions(transaction, transactions, orderOfQuery) ||
		checker.IsDifferentLocations(lastLocation, current)
}

func (d *DefaultTransactionChecker) IsAboveLimit(transaction transaction.Transaction) bool {
	if transaction.Amount <= 0 {
		return false
	}
	return transaction.Amount > MaxAmount
}

func (d *DefaultTransactionChecker) OccurMultipleTransactions(
	transactionReceived transaction.Transaction,
	transactions []transaction.Transaction,
	orderOfQuery string,
) bool {
	if len(transactions) == 0 {
		return false
	}

	var lastTransaction transaction.Transaction
	if orderOfQuery == "DESC" {
		lastTransaction = transactions[0]
	} else {
		lastTransaction = transactions[len(transactions)-1]
	}

	timeDiff := transactionReceived.TransactionTime.UTC().Sub(lastTransaction.TransactionTime.UTC())
	return timeDiff < 30*time.Second
}

func (d *DefaultTransactionChecker) IsDifferentLocations(last string, current string) bool {
	if last == "" || current == "" {
		return false
	}

	return last != current
}
