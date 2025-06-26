package test

import (
	"fraud-detection/internal/frauddetector"
	"fraud-detection/pkg/fraud"
	"fraud-detection/pkg/transaction"
	"testing"
	"time"
)

var checker = &frauddetector.DefaultTransactionChecker{}

func TestOccurMultipleTransactions(t *testing.T) {
	tests := []struct {
		name           string
		transac        transaction.Transaction
		transactions   []transaction.Transaction
		expectedResult bool
	}{
		{
			name: "Suspicious transaction within 30 seconds",
			transac: transaction.Transaction{
				ID:        "t1",
				Amount:    100.50,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				AccountID: "123",
				Location:  "New York",
				IPAddress: "192.168.1.1",
			},
			transactions: []transaction.Transaction{
				{
					ID:        "t2",
					Amount:    50.00,
					Timestamp: time.Now().Add(-15 * time.Second).Format("2006-01-02 15:04:05"),
					AccountID: "123",
					Location:  "New York",
					IPAddress: "192.168.1.1",
				},
			},
			expectedResult: true,
		},
		{
			name: "Not suspicious transaction (over 30 seconds)",
			transac: transaction.Transaction{
				ID:        "t1",
				Amount:    100.50,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				AccountID: "123",
				Location:  "New York",
				IPAddress: "192.168.1.1",
			},
			transactions: []transaction.Transaction{
				{
					ID:        "t2",
					Amount:    50.00,
					Timestamp: time.Now().Add(-45 * time.Second).Format("2006-01-02 15:04:05"),
					AccountID: "123",
					Location:  "New York",
					IPAddress: "192.168.1.1",
				},
			},
			expectedResult: false,
		},
		{
			name: "No previous transaction for account",
			transac: transaction.Transaction{
				ID:        "t1",
				Amount:    100.50,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				AccountID: "123",
				Location:  "New York",
				IPAddress: "192.168.1.1",
			},
			transactions:   []transaction.Transaction{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.OccurMultipleTransactions(tt.transac, tt.transactions)
			if result != tt.expectedResult {
				t.Errorf("OccurMultipleTransactions() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestIsAboveLimit(t *testing.T) {
	tests := []struct {
		name           string
		transaction    transaction.Transaction
		expectedResult bool
	}{
		{
			name: "Transaction amount greater than MaxAmount",
			transaction: transaction.Transaction{
				Amount: frauddetector.MaxAmount + 1,
			},
			expectedResult: true,
		},
		{
			name: "Transaction amount equal to MaxAmount",
			transaction: transaction.Transaction{
				Amount: 1000.0,
			},
			expectedResult: false,
		},
		{
			name: "Transaction amount less than MaxAmount",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			expectedResult: false,
		},
		{
			name: "Transaction amount negative (invalid scenario)",
			transaction: transaction.Transaction{
				Amount: -100.0,
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.IsAboveLimit(tt.transaction)
			if result != tt.expectedResult {
				t.Errorf("isAboveLimit() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

type MockTransactionChecker struct {
	IsAboveLimitFunc              func(transaction transaction.Transaction) bool
	OccurMultipleTransactionsFunc func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool
}

func (m *MockTransactionChecker) IsAboveLimit(transaction transaction.Transaction) bool {
	return m.IsAboveLimitFunc(transaction)
}

func (m *MockTransactionChecker) OccurMultipleTransactions(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
	return m.OccurMultipleTransactionsFunc(transactionReceived, transactions)
}

func TestIsSuspiciousTransaction(t *testing.T) {
	tests := []struct {
		name           string
		transaction    transaction.Transaction
		transactions   []transaction.Transaction
		mockChecker    *MockTransactionChecker
		expectedResult bool
	}{
		{
			name: "Suspicious transaction (above limit)",
			transaction: transaction.Transaction{
				Amount: 1500000,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return true
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return false
				},
			},
			expectedResult: true,
		},
		{
			name: "Suspicious transaction (multiple transactions)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{
				{
					Amount: 300.0,
				},
				{
					Amount: 600.0,
				},
			},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return false
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return true
				},
			},
			expectedResult: true,
		},
		{
			name: "Not suspicious transaction (below limit)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return false
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return false
				},
			},
			expectedResult: false,
		},
		{
			name: "Suspicious transaction (Above limit and Multiple Transactions)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return true
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return true
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frauddetector.IsSuspiciousTransaction(tt.transaction, tt.transactions, tt.mockChecker)
			if result != tt.expectedResult {
				t.Errorf("IsSuspiciousTransaction() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestCheckTransactionForFraud(t *testing.T) {
	tests := []struct {
		name           string
		transaction    transaction.Transaction
		transactions   []transaction.Transaction
		mockChecker    *MockTransactionChecker
		expectedResult bool
	}{
		{
			name: "Suspicious transaction (above limit)",
			transaction: transaction.Transaction{
				Amount: 1500000,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return true
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return false
				},
			},
			expectedResult: true,
		},
		{
			name: "Suspicious transaction (multiple transactions)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{
				{
					Amount: 300.0,
				},
				{
					Amount: 600.0,
				},
			},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return false
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return true
				},
			},
			expectedResult: true,
		},
		{
			name: "Not suspicious transaction (below limit)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return false
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return false
				},
			},
			expectedResult: false,
		},
		{
			name: "Suspicious transaction (Above limit and Multiple Transactions)",
			transaction: transaction.Transaction{
				Amount: 500.0,
			},
			transactions: []transaction.Transaction{},
			mockChecker: &MockTransactionChecker{
				IsAboveLimitFunc: func(transaction transaction.Transaction) bool {
					return true
				},
				OccurMultipleTransactionsFunc: func(transactionReceived transaction.Transaction, transactions []transaction.Transaction) bool {
					return true
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fraud.CheckTransactionForFraud(tt.transaction, tt.transactions, tt.mockChecker)
			if result != tt.expectedResult {
				t.Errorf("IsSuspiciousTransaction() = %v; want %v", result, tt.expectedResult)
			}
		})
	}
}
