package frauddetector

import (
	"fraud-detection/pkg/transaction"
	"time"
)

const MaxAmount = 1000000

func IsSuspiciousTransaction(
	transaction transaction.Transaction,
	transactions []transaction.Transaction,
	checker TransactionChecker,
) bool {
	return checker.IsAboveLimit(transaction) || checker.OccurMultipleTransactions(transaction, transactions)
}

func (d *DefaultTransactionChecker) IsAboveLimit(transaction transaction.Transaction) bool {
	return transaction.Amount > MaxAmount
}

func (d *DefaultTransactionChecker) OccurMultipleTransactions(
	transactionReceived transaction.Transaction,
	transactions []transaction.Transaction,
) bool {
	// filter transactions (db, memory) for the account_id
	var lastTransaction transaction.Transaction

	// get last transaction from this filter
	if len(transactions) > 0 {
		lastTransaction = transactions[len(transactions)-1]
	}

	duration := transactionReceived.Timestamp.UTC().Sub(lastTransaction.Timestamp.UTC())
	if duration < 30*time.Second {
		// is a suspicious transaction
		return true
	} else {
		// no is a suspicious transaction
		return false
	}
}
