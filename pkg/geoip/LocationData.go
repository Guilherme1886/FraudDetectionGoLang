package geoip

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LocationData struct {
	Country   string  `json:"country_name"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func GetGeoLocation(ip string) (*LocationData, error) {
	const apiKey = "b0ba45a6129dbd88f5d1aafb70ca22af"
	if apiKey == "" {
		return nil, fmt.Errorf("API key not provided")
	}

	url := fmt.Sprintf("https://api.ipstack.com/%s?access_key=%s", ip, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch geolocation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var location LocationData
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &location, nil
}
