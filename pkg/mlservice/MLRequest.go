package mlservice

type MLRequest struct {
	Features []float64 `json:"features"`
}
