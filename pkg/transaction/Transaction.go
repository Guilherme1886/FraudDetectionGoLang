package transaction

import "time"

type Transaction struct {
	ID        string  `json:"id"`
	Amount    float64 `json:"amount"`
	Timestamp string  `json:"timestamp"`
	AccountID string  `json:"account_id"`
	Location  string  `json:"location"`
	IPAddress string  `json:"ip_address"`
}

func NewTransaction(id string, amount float64, accountID, location, ipAddress string) Transaction {
	return Transaction{
		ID:        id,
		Amount:    amount,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		AccountID: accountID,
		Location:  location,
		IPAddress: ipAddress,
	}
}
