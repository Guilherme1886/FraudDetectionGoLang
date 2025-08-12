package fraud

import (
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/transaction"
)

func CheckTransactionForFraud(
	transaction transaction.Transaction,
	history []transaction.Transaction,
	previousLocation string,
	currentLocation string,
	orderOfQuery string,
	checker frauddetector.TransactionChecker,
) bool {
	return frauddetector.IsSuspiciousTransaction(
		transaction,
		history,
		previousLocation,
		currentLocation,
		orderOfQuery,
		checker,
	)
}
