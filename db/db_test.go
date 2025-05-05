package db

import (
	"os"
	"testing"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

func TestNew(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Test creating a new database
	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify the database connection works
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new database
	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test initializing the database
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify the transactions table was created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='transactions'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to query for transactions table: %v", err)
	}
	if tableName != "transactions" {
		t.Fatalf("Expected table name 'transactions', got '%s'", tableName)
	}
}

func TestSaveAndGetTransaction(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new database
	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a test transaction
	tx := &models.Transaction{
		ReferenceNumber: "TEST123",
		ActivityType:    "TRANS",
		Amount: &models.Amount{
			Value:    "25.99",
			Currency: "USD",
		},
		ActivityStatus:         "APPROVED",
		ActivityCategory:       "PURCHASE",
		ActivityClassification: "PURCHASE",
		CardNumber:             "************1234",
		Merchant: &models.Merchant{
			Name:     "Test Merchant",
			Category: "RETAIL",
			Address: &models.Address{
				City:          "Test City",
				StateProvince: "TS",
				PostalCode:    "12345",
				CountryCode:   "US",
			},
		},
		Date:                 "2025-04-29",
		ActivityCategoryCode: "0001",
		CustomerID:           "TEST",
		PostedDate:           "2025-04-29",
		Name:                 &models.Name{NameOnCard: "TEST USER"},
	}

	// Test saving the transaction
	if err := db.SaveTransaction(tx); err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Test retrieving the transaction
	retrievedTx, err := db.GetTransactionByReference("TEST123")
	if err != nil {
		t.Fatalf("Failed to retrieve transaction: %v", err)
	}

	// Verify the transaction was retrieved correctly
	if retrievedTx.ReferenceNumber != tx.ReferenceNumber {
		t.Errorf("Expected reference number '%s', got '%s'", tx.ReferenceNumber, retrievedTx.ReferenceNumber)
	}
	if retrievedTx.Amount.Value != tx.Amount.Value {
		t.Errorf("Expected amount value '%s', got '%s'", tx.Amount.Value, retrievedTx.Amount.Value)
	}
	if retrievedTx.Amount.Currency != tx.Amount.Currency {
		t.Errorf("Expected amount currency '%s', got '%s'", tx.Amount.Currency, retrievedTx.Amount.Currency)
	}
	if retrievedTx.Merchant.Name != tx.Merchant.Name {
		t.Errorf("Expected merchant name '%s', got '%s'", tx.Merchant.Name, retrievedTx.Merchant.Name)
	}
}

func TestUpdateTransaction(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new database
	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a test transaction
	tx := &models.Transaction{
		ReferenceNumber: "TEST123",
		ActivityType:    "TRANS",
		Amount: &models.Amount{
			Value:    "25.99",
			Currency: "USD",
		},
		ActivityStatus:         "APPROVED",
		ActivityCategory:       "PURCHASE",
		ActivityClassification: "PURCHASE",
		CardNumber:             "************1234",
		Merchant: &models.Merchant{
			Name:     "Test Merchant",
			Category: "RETAIL",
			Address: &models.Address{
				City:          "Test City",
				StateProvince: "TS",
				PostalCode:    "12345",
				CountryCode:   "US",
			},
		},
		Date:                 "2025-04-29",
		ActivityCategoryCode: "0001",
		CustomerID:           "TEST",
		PostedDate:           "2025-04-29",
		Name:                 &models.Name{NameOnCard: "TEST USER"},
	}

	// Save the transaction
	if err := db.SaveTransaction(tx); err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Update the transaction
	tx.Amount.Value = "30.99"
	tx.Merchant.Name = "Updated Merchant"
	tx.LunchMoneyID = 12345

	// Test updating the transaction
	if err := db.UpdateTransaction(tx); err != nil {
		t.Fatalf("Failed to update transaction: %v", err)
	}

	// Retrieve the updated transaction
	retrievedTx, err := db.GetTransactionByReference("TEST123")
	if err != nil {
		t.Fatalf("Failed to retrieve updated transaction: %v", err)
	}

	// Verify the transaction was updated correctly
	if retrievedTx.Amount.Value != "30.99" {
		t.Errorf("Expected updated amount value '30.99', got '%s'", retrievedTx.Amount.Value)
	}
	if retrievedTx.Merchant.Name != "Updated Merchant" {
		t.Errorf("Expected updated merchant name 'Updated Merchant', got '%s'", retrievedTx.Merchant.Name)
	}
	if retrievedTx.LunchMoneyID != 12345 {
		t.Errorf("Expected LunchMoneyID 12345, got %d", retrievedTx.LunchMoneyID)
	}
}

func TestRemoveTransaction(t *testing.T) {
	// Create a temporary database file
	tempFile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create a new database
	db, err := New(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize the database
	if err := db.Initialize(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a test transaction
	tx := &models.Transaction{
		ReferenceNumber: "TEST123",
		ActivityType:    "TRANS",
		Amount: &models.Amount{
			Value:    "25.99",
			Currency: "USD",
		},
		ActivityStatus:         "APPROVED",
		ActivityCategory:       "PURCHASE",
		ActivityClassification: "PURCHASE",
		CardNumber:             "************1234",
		Merchant: &models.Merchant{
			Name:     "Test Merchant",
			Category: "RETAIL",
			Address: &models.Address{
				City:          "Test City",
				StateProvince: "TS",
				PostalCode:    "12345",
				CountryCode:   "US",
			},
		},
		Date:                 "2025-04-29",
		ActivityCategoryCode: "0001",
		CustomerID:           "TEST",
		PostedDate:           "2025-04-29",
		Name:                 &models.Name{NameOnCard: "TEST USER"},
	}

	// Save the transaction
	if err := db.SaveTransaction(tx); err != nil {
		t.Fatalf("Failed to save transaction: %v", err)
	}

	// Test removing the transaction
	if err := db.RemoveTransaction("TEST123"); err != nil {
		t.Fatalf("Failed to remove transaction: %v", err)
	}

	// Verify the transaction was removed
	retrievedTx, err := db.GetTransactionByReference("TEST123")
	if err != nil {
		t.Fatalf("Error when checking if transaction was removed: %v", err)
	}
	if retrievedTx != nil {
		t.Errorf("Expected transaction to be removed, but it still exists")
	}
}
