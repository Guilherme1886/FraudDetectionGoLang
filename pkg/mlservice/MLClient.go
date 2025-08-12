package mlservice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

var baseURL = "http://127.0.0.1:5000"

type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient() *MLClient {
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *MLClient) PredictFraud(features []float64) (*MLResponse, error) {
	request := MLRequest{
		Features: features,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/predict", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prediction MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, err
	}

	return &prediction, nil
}
