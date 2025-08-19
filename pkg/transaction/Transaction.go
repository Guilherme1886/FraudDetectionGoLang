package transaction

import (
	"github.com/google/uuid"
	"time"
)

type Label int

const (
	LabelLegit Label = iota
	LabelFraud
)

type Transaction struct {
	TransactionID   string    `json:"transaction_id"`
	Amount          float64   `json:"amount"`
	AccountID       string    `json:"account_id"`
	Location        string    `json:"location"`
	TransactionTime time.Time `json:"transaction_time"`
	ElapsedTime     float64   `json:"elapsed_time"`
	Frequency       int       `json:"frequency"`
	FraudLabel      Label     `json:"fraud_label"`
}

func NewTransaction(amount float64, accountID, location string, elapsedTime float64, frequency int) Transaction {
	if amount < 0 {
		amount = 0
	}
	return Transaction{
		TransactionID:   uuid.New().String(),
		Amount:          amount,
		AccountID:       accountID,
		Location:        location,
		TransactionTime: time.Now().UTC(),
		ElapsedTime:     elapsedTime,
		Frequency:       frequency,
		FraudLabel:      LabelLegit,
	}
}
