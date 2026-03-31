package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ArcadeDBClient handles HTTP communication with ArcadeDB.
type ArcadeDBClient struct {
	baseURL    string
	database   string
	username   string
	password   string
	httpClient *http.Client
}

// NewArcadeDBClient creates a new ArcadeDB HTTP client.
func NewArcadeDBClient(baseURL, database, username, password string) *ArcadeDBClient {
	return &ArcadeDBClient{
		baseURL:  baseURL,
		database: database,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckHealth verifies the connection to ArcadeDB.
func (c *ArcadeDBClient) CheckHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/ts/%s/grafana/health", c.baseURL, c.database)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetMetadata fetches time series metadata from ArcadeDB.
func (c *ArcadeDBClient) GetMetadata(ctx context.Context) (*MetadataResponse, error) {
	url := fmt.Sprintf("%s/api/v1/ts/%s/grafana/metadata", c.baseURL, c.database)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("metadata request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("metadata returned status %d: %s", resp.StatusCode, string(body))
	}

	var metadata MetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

// QueryTimeSeries sends a query to the Grafana query endpoint.
func (c *ArcadeDBClient) QueryTimeSeries(ctx context.Context, request *GrafanaQueryRequest) (map[string]json.RawMessage, error) {
	url := fmt.Sprintf("%s/api/v1/ts/%s/grafana/query", c.baseURL, c.database)

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Results map[string]json.RawMessage `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	return result.Results, nil
}

// ExecuteCommand sends a command to ArcadeDB and returns the raw response.
func (c *ArcadeDBClient) ExecuteCommand(ctx context.Context, req *CommandRequest) ([]byte, error) {
	url := fmt.Sprintf("%s/api/v1/command/%s", c.baseURL, c.database)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("command request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("command returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

// doRequest executes an HTTP request with authentication.
func (c *ArcadeDBClient) doRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}
