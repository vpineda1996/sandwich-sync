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

	// Verify the account mappings table was created
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='account_mappings'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Failed to query for account_mappings table: %v", err)
	}
	if tableName != "account_mappings" {
		t.Fatalf("Expected table name 'account_mappings', got '%s'", tableName)
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
	tx := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TEST123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name: "Test Merchant",
				Address: &models.Address{
					City:          "Test City",
					StateProvince: "TS",
				},
			},
			Date:       "2025-04-29",
			PostedDate: "2025-04-29",
		},
		SourceAccountName: "Test Account",
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
	if retrievedTx.SourceAccountName != tx.SourceAccountName {
		t.Errorf("Expected source account name '%s', got '%s'", tx.SourceAccountName, retrievedTx.SourceAccountName)
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
	tx := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TEST123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name: "Test Merchant",
				Address: &models.Address{
					City:          "Test City",
					StateProvince: "TS",
				},
			},
			Date:       "2025-04-29",
			PostedDate: "2025-04-29",
		},
		SourceAccountName: "Test Account",
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
	if retrievedTx.SourceAccountName != tx.SourceAccountName {
		t.Errorf("Expected source account name '%s', got '%s'", tx.SourceAccountName, retrievedTx.SourceAccountName)
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
	tx := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TEST123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name: "Test Merchant",
				Address: &models.Address{
					City:          "Test City",
					StateProvince: "TS",
				},
			},
			Date:       "2025-04-29",
			PostedDate: "2025-04-29",
		},
		SourceAccountName: "Test Account",
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
func TestSaveAndGetAccountMapping(t *testing.T) {
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

	// Create a test account mapping
	accountMapping := &models.AccountMapping{
		ExternalName: "Test External Account",
		LunchMoneyId: 12345,
		IsPlaid:      true,
	}

	// Test saving the account mapping
	if err := db.UpsertAccountMapping(accountMapping); err != nil {
		t.Fatalf("Failed to save account mapping: %v", err)
	}

	// Test retrieving the account mapping
	retrievedMapping, err := db.GetAccountMapping("Test External Account")
	if err != nil {
		t.Fatalf("Failed to retrieve account mapping: %v", err)
	}

	// Verify the account mapping was retrieved correctly
	if retrievedMapping.ExternalName != accountMapping.ExternalName {
		t.Errorf("Expected external name '%s', got '%s'", accountMapping.ExternalName, retrievedMapping.ExternalName)
	}
	if retrievedMapping.LunchMoneyId != accountMapping.LunchMoneyId {
		t.Errorf("Expected LunchMoneyId '%d', got '%d'", accountMapping.LunchMoneyId, retrievedMapping.LunchMoneyId)
	}
	if retrievedMapping.IsPlaid != accountMapping.IsPlaid {
		t.Errorf("Expected IsPlaid '%v', got '%v'", accountMapping.IsPlaid, retrievedMapping.IsPlaid)
	}
}
