package homeassistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HAClient is a client for Home Assistant REST API
type HAClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// HAState represents the state of an entity in Home Assistant
type HAState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
}

// HAAPIStatus represents the API status response
type HAAPIStatus struct {
	Message string `json:"message"`
}

// NewHAClient creates a new Home Assistant client
func NewHAClient(baseURL, token string) *HAClient {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	return &HAClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to Home Assistant API
func (c *HAClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// CheckConnection verifies the connection to Home Assistant
func (c *HAClient) CheckConnection(ctx context.Context) error {
	body, err := c.doRequest(ctx, http.MethodGet, "/api/", nil)
	if err != nil {
		return err
	}

	var status HAAPIStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse API status: %w", err)
	}

	if status.Message != "API running." {
		return fmt.Errorf("unexpected API status: %s", status.Message)
	}

	return nil
}

// GetStates returns all entity states from Home Assistant
func (c *HAClient) GetStates(ctx context.Context) ([]HAState, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/api/states", nil)
	if err != nil {
		return nil, err
	}

	var states []HAState
	if err := json.Unmarshal(body, &states); err != nil {
		return nil, fmt.Errorf("failed to parse states: %w", err)
	}

	return states, nil
}

// GetState returns the state of a specific entity
func (c *HAClient) GetState(ctx context.Context, entityID string) (*HAState, error) {
	body, err := c.doRequest(ctx, http.MethodGet, "/api/states/"+entityID, nil)
	if err != nil {
		return nil, err
	}

	var state HAState
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

// CallService calls a service in Home Assistant
func (c *HAClient) CallService(ctx context.Context, domain, service string, data map[string]interface{}) error {
	endpoint := fmt.Sprintf("/api/services/%s/%s", domain, service)

	_, err := c.doRequest(ctx, http.MethodPost, endpoint, data)
	return err
}

// CallServiceForEntity calls a service for a specific entity
func (c *HAClient) CallServiceForEntity(ctx context.Context, entityID, service string, data map[string]interface{}) error {
	// Extract domain from entity_id (e.g., "light.living_room" -> "light")
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid entity_id format: %s", entityID)
	}
	domain := parts[0]

	// Add entity_id to data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["entity_id"] = entityID

	return c.CallService(ctx, domain, service, data)
}

// GetEntityDomain extracts the domain from an entity ID
func GetEntityDomain(entityID string) string {
	parts := strings.SplitN(entityID, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetFriendlyName returns the friendly name from attributes or entity_id
func (s *HAState) GetFriendlyName() string {
	if name, ok := s.Attributes["friendly_name"].(string); ok {
		return name
	}
	return s.EntityID
}

// GetDeviceClass returns the device_class attribute if present
func (s *HAState) GetDeviceClass() string {
	if class, ok := s.Attributes["device_class"].(string); ok {
		return class
	}
	return ""
}

// IsAvailable returns true if the entity is not unavailable
func (s *HAState) IsAvailable() bool {
	return s.State != "unavailable" && s.State != "unknown"
}
