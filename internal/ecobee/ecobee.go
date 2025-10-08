package ecobee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles fetching temperature from Ecobee via homebridge
type Client struct {
	baseURL    string
	deviceName string
	client     *http.Client
	enabled    bool
}

// TemperatureResponse represents the homebridge webhook response
type TemperatureResponse struct {
	CurrentTemperature float64 `json:"currentTemperature"`
}

// New creates a new Ecobee client
// If baseURL is empty, client is disabled and will return 0 for all temps
func New(baseURL, deviceName string) *Client {
	enabled := baseURL != "" && deviceName != ""

	return &Client{
		baseURL:    baseURL,
		deviceName: deviceName,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		enabled: enabled,
	}
}

// IsEnabled returns whether the Ecobee integration is configured
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// GetTemperature fetches the current temperature from Ecobee
// Returns 0 if disabled or on error (caller should handle gracefully)
func (c *Client) GetTemperature() (float64, error) {
	if !c.enabled {
		return 0, nil
	}

	// Construct URL - format depends on homebridge plugin
	// For homebridge-http-webhooks: http://host:port/device/deviceName
	url := fmt.Sprintf("%s/%s", c.baseURL, c.deviceName)

	resp, err := c.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch temperature: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var tempResp TemperatureResponse
	if err := json.Unmarshal(body, &tempResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert Celsius to Fahrenheit if needed
	// Homebridge usually returns Celsius for temperature sensors
	tempF := celsiusToFahrenheit(tempResp.CurrentTemperature)

	return tempF, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}
