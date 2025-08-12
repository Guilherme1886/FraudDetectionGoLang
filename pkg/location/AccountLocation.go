package location

type AccountLocation struct {
	ID        int64  `json:"id"`
	AccountID string `json:"account_id"`
	City      string `json:"city"`
}
