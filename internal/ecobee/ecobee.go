package ecobee

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles fetching temperature from Ecobee via Home Assistant
type Client struct {
	baseURL    string
	token      string
	entityID   string
	client     *http.Client
	enabled    bool
}

// HAStateResponse represents the Home Assistant API state response
type HAStateResponse struct {
	State string `json:"state"`
}

// New creates a new Ecobee client
// baseURL: Home Assistant base URL (e.g., "http://localhost:8123")
// token: Home Assistant long-lived access token
// entityID: Ecobee temperature sensor entity ID (e.g., "sensor.my_ecobee_current_temperature")
func New(baseURL, token, entityID string) *Client {
	enabled := baseURL != "" && token != "" && entityID != ""

	return &Client{
		baseURL:  baseURL,
		token:    token,
		entityID: entityID,
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

// GetTemperature fetches the current temperature from Ecobee via Home Assistant
// Returns 0 if disabled or on error (caller should handle gracefully)
func (c *Client) GetTemperature() (float64, error) {
	if !c.enabled {
		return 0, nil
	}

	// Construct URL for Home Assistant API
	url := fmt.Sprintf("%s/api/states/%s", c.baseURL, c.entityID)

	// Create request with authorization header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
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

	var stateResp HAStateResponse
	if err := json.Unmarshal(body, &stateResp); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse temperature from state string
	var temp float64
	if _, err := fmt.Sscanf(stateResp.State, "%f", &temp); err != nil {
		return 0, fmt.Errorf("failed to parse temperature: %w", err)
	}

	// Temperature is already in Fahrenheit from Home Assistant
	return temp, nil
}
