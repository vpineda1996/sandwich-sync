package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

func TestExtractCookies(t *testing.T) {
	testCases := []struct {
		name            string
		cookieHeader    string
		expectedCookies map[string]string
	}{
		{
			name:            "Single cookie",
			cookieHeader:    "name=value",
			expectedCookies: map[string]string{"name": "value"},
		},
		{
			name:         "Multiple cookies",
			cookieHeader: "name1=value1; name2=value2; name3=value3",
			expectedCookies: map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			},
		},
		{
			name:         "Cookies with spaces",
			cookieHeader: " name1 = value1 ; name2= value2;name3 =value3 ",
			expectedCookies: map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			},
		},
		{
			name:            "Empty cookie header",
			cookieHeader:    "",
			expectedCookies: map[string]string{},
		},
		{
			name:            "Invalid cookie format",
			cookieHeader:    "invalid; format; missing=equals",
			expectedCookies: map[string]string{"missing": "equals"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractCookies(tc.cookieHeader)

			// Check if the maps have the same length
			if len(result) != len(tc.expectedCookies) {
				t.Errorf("Expected %d cookies, got %d", len(tc.expectedCookies), len(result))
			}

			// Check if all expected cookies are present with correct values
			for key, expectedValue := range tc.expectedCookies {
				if value, ok := result[key]; !ok {
					t.Errorf("Expected cookie '%s' not found", key)
				} else if value != expectedValue {
					t.Errorf("Expected cookie '%s' to have value '%s', got '%s'", key, expectedValue, value)
				}
			}
		})
	}
}

func TestFetchTransactions(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
		}

		// Create a sample response
		transactions := []*models.Transaction{
			{
				ReferenceNumber: "TX123",
				ActivityType:    "TRANS",
				Amount: &models.Amount{
					Value:    "25.99",
					Currency: "USD",
				},
				Merchant: &models.Merchant{
					Name: "Test Merchant",
				},
				Date: "2025-04-29",
			},
		}

		response := activitiesResponse{
			Activities: transactions,
		}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a client
	client := NewCurlClient()

	// Set up headers
	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"Content-Type":  "application/json",
	}

	// Fetch transactions
	transactions, err := client.FetchTransactions(server.URL, headers)
	if err != nil {
		t.Fatalf("Failed to fetch transactions: %v", err)
	}

	// Verify the response
	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.ReferenceNumber != "TX123" {
		t.Errorf("Expected reference number 'TX123', got '%s'", tx.ReferenceNumber)
	}
	if tx.Amount.Value != "25.99" {
		t.Errorf("Expected amount value '25.99', got '%s'", tx.Amount.Value)
	}
	if tx.Amount.Currency != "USD" {
		t.Errorf("Expected amount currency 'USD', got '%s'", tx.Amount.Currency)
	}
	if tx.Merchant.Name != "Test Merchant" {
		t.Errorf("Expected merchant name 'Test Merchant', got '%s'", tx.Merchant.Name)
	}
}

func TestFetchTransactionsError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	// Create a client
	client := NewCurlClient()

	// Set up headers
	headers := map[string]string{
		"Authorization": "Bearer invalid-token",
	}

	// Fetch transactions (should fail)
	_, err := client.FetchTransactions(server.URL, headers)
	if err == nil {
		t.Errorf("Expected error when server returns unauthorized, got nil")
	}
}
