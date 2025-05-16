package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func TestCreateAccountInfoTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// The table should be created during initialization, but we can test the method directly
	err := db.createAccountInfoTable()
	assert.NoError(t, err)

	// Verify table exists by inserting a record
	_, err = db.Exec("INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, is_plaid) VALUES (?, ?, ?, ?)",
		4, "100.00", "CAD", false)
	assert.NoError(t, err)
}

func setupTestDB(t *testing.T) *DB {
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return db
}

func TestUpsertAccountBalance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// First create the necessary account mapping
	_, err := db.Exec("INSERT INTO account_mappings (lunchmoney_account_id, external_name) VALUES (?, ?)",
		"12", "test-external-name")
	assert.NoError(t, err)

	// Test upserting balance
	err = db.UpsertAccountBalance("test-external-name", models.Amount{
		Value:    "200.50",
		Currency: "CAD",
	})
	assert.NoError(t, err)

	// Verify the balance was saved correctly
	var value, currency string
	var updatedAt time.Time
	err = db.QueryRow("SELECT balance_value, balance_currency, balance_updated_at FROM account_info WHERE lunchmoney_account_id = ?",
		12).Scan(&value, &currency, &updatedAt)
	assert.NoError(t, err)
	assert.Equal(t, "200.50", value)
	assert.Equal(t, "CAD", currency)
	assert.NotEmpty(t, updatedAt)
}

func TestGetAccounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Setup test data
	now := time.Now().UTC().Truncate(time.Second)

	_, err := db.Exec("INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, balance_updated_at, is_plaid) VALUES (?, ?, ?, ?, false)",
		1, "150.75", "EUR", now)
	assert.NoError(t, err)

	// Get accounts
	accounts, err := db.GetAccounts()
	assert.NoError(t, err)

	// Verify returned data
	assert.Len(t, accounts, 1)
	if len(accounts) > 0 {
		assert.Equal(t, int64(1), accounts[0].LunchMoneyId)
		assert.Equal(t, "150.75", accounts[0].Balance.Value)
		assert.Equal(t, "EUR", accounts[0].Balance.Currency)
		assert.False(t, accounts[0].IsPlaid)
		assert.WithinDuration(t, now, *accounts[0].BalanceLastUpdated, time.Second)
	}
}
func TestDisableAccountSync(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert an account that has sync enabled by default
	_, err := db.Exec("INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, should_sync) VALUES (?, ?, ?, ?)",
		42, "300.00", "USD", true)
	assert.NoError(t, err)

	// Verify initial state
	var shouldSync bool
	err = db.QueryRow("SELECT should_sync FROM account_info WHERE lunchmoney_account_id = ?", 42).Scan(&shouldSync)
	assert.NoError(t, err)
	assert.True(t, shouldSync)

	// Disable sync
	err = db.DisableAccountSync("42")
	assert.NoError(t, err)

	// Verify sync was disabled
	err = db.QueryRow("SELECT should_sync FROM account_info WHERE lunchmoney_account_id = ?", 42).Scan(&shouldSync)
	assert.NoError(t, err)
	assert.False(t, shouldSync)
}

func TestIsAccountSyncEnabled(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test 1: Account exists with sync enabled
	_, err := db.Exec("INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, should_sync) VALUES (?, ?, ?, ?)",
		100, "500.00", "USD", true)
	assert.NoError(t, err)

	// Test 2: Account exists with sync disabled
	_, err = db.Exec("INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, should_sync) VALUES (?, ?, ?, ?)",
		101, "750.00", "USD", false)
	assert.NoError(t, err)

	// Test cases
	testCases := []struct {
		name          string
		accountID     int64
		expectedSync  bool
		expectedError bool
	}{
		{
			name:          "Account with sync enabled",
			accountID:     100,
			expectedSync:  true,
			expectedError: false,
		},
		{
			name:          "Account with sync disabled",
			accountID:     101,
			expectedSync:  false,
			expectedError: false,
		},
		{
			name:          "Non-existent account",
			accountID:     999,
			expectedSync:  false,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			syncEnabled, err := db.IsAccountSyncEnabled(tc.accountID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedSync, syncEnabled)
		})
	}
}
