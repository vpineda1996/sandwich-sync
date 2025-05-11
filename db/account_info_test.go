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
