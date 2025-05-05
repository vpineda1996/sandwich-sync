package rogers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

// CurlClient represents an HTTP client for making API requests
type CurlClient struct {
	httpClient *http.Client
}

// NewCurlClient creates a new HTTP client
func NewCurlClient() *CurlClient {
	return &CurlClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type activitiesResponse struct {
	Activities []models.Transaction `json:"activities"`
}

// FetchTransactions fetches transactions from the API
func (c *CurlClient) FetchTransactions(url string, headers map[string]string) ([]models.Transaction, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var transactions activitiesResponse
	if err := json.Unmarshal(body, &transactions); err != nil {
		// Try to print the response body for debugging
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	return transactions.Activities, nil
}

// ExtractCookies extracts cookies from a cookie header string
func ExtractCookies(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	parts := strings.Split(cookieHeader, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		cookies[key] = value
	}

	return cookies
}
