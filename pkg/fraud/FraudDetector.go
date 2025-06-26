package fraud

import (
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/transaction"
)

func CheckTransactionForFraud(t transaction.Transaction, transactions []transaction.Transaction, checker frauddetector.TransactionChecker) bool {
	return frauddetector.IsSuspiciousTransaction(t, transactions, checker)
}
