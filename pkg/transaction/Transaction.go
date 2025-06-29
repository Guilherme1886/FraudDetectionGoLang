package transaction

import "time"

type Transaction struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	AccountID string    `json:"account_id"`
	Location  string    `json:"location"`
	IPAddress string    `json:"ip_address"`
}

func NewTransaction(amount float64, accountID, location, ipAddress string) Transaction {
	return Transaction{
		Amount:    amount,
		Timestamp: time.Now().UTC(),
		AccountID: accountID,
		Location:  location,
		IPAddress: ipAddress,
	}
}
