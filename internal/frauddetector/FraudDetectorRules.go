package frauddetector

import (
	"fmt"
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
	var filter []transaction.Transaction
	var lastTransaction transaction.Transaction

	for _, tr := range transactions {
		if tr.AccountID == transactionReceived.AccountID {
			filter = append(filter, tr)
		}
	}
	// get last transaction from this filter
	if len(filter) > 0 {
		lastTransaction = filter[len(filter)-1]
	}

	// validate between the two timestamps if the transactions occur in less of 30 seconds
	layout := "2006-01-02 15:04:05"
	t1, err := time.Parse(layout, transactionReceived.Timestamp)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
	}

	t2, err := time.Parse(layout, lastTransaction.Timestamp)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
	}

	duration := t1.Sub(t2)
	if duration < 30*time.Second {
		// is a suspicious transaction
		return true
	} else {
		// no is a suspicious transaction
		return false
	}
}
