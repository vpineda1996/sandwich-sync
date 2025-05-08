package services

import (
	"context"
	"testing"
	"time"

	"github.com/vpnda/sandwich-sync/db"
	"github.com/vpnda/sandwich-sync/pkg/http/lm"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func TestSyncTransactions(t *testing.T) {
	// Create mock database
	mockDB := db.NewMockDB()

	// Add some test transactions to the database
	tx1 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 1",
				Address: &models.Address{},
			},
			Date:         time.Now().Format(time.DateOnly),
			LunchMoneyID: 0, // Not synced yet
		},
		SourceAccountName: "Test Account",
	}

	tx2 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX456",
			Amount: models.Amount{
				Value:    "50.00",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 2",
				Address: &models.Address{},
			},
			Date:         time.Now().Format(time.DateOnly),
			LunchMoneyID: 12345, // Already synced
		},
		SourceAccountName: "Test Account",
	}

	mockDB.Transactions[tx1.ReferenceNumber] = tx1
	mockDB.Transactions[tx2.ReferenceNumber] = tx2

	// Create mock LunchMoney client
	mockClient := &lm.MockLunchMoneyClient{
		Accounts: []models.LunchMoneyAccount{
			{LunchMoneyId: 1, Name: "Test Account"},
		},
		Transactions: []models.Transaction{
			{
				ReferenceNumber: "TX456",
				Amount: models.Amount{
					Value:    "50.00",
					Currency: "USD",
				},
				Merchant: &models.Merchant{
					Name: "Test Merchant 2",
				},
				Date:         time.Now().Format(time.DateOnly),
				LunchMoneyID: 12345,
			},
		},
		InsertedIDs: []int64{54321},
	}

	// Create a mock account selector
	mockSelector := NewAccountSelectorWithClient(mockClient, mockDB)
	mockSelector.selectedAccount = &models.AccountMapping{
		LunchMoneyId: 1,
		ExternalName: "Test Account",
	}

	// Create the syncer
	syncer := &LunchMoneySyncer{
		client:          mockClient,
		database:        mockDB,
		accountSelector: mockSelector,
	}

	// Test syncing transactions
	err := syncer.SyncTransactions(context.Background())
	if err != nil {
		t.Fatalf("Failed to sync transactions: %v", err)
	}

	// Verify that TX123 was synced
	syncedTx1, err := mockDB.GetTransactionByReference("TX123")
	if err != nil {
		t.Fatalf("Failed to get transaction TX123: %v", err)
	}
	if syncedTx1.LunchMoneyID != 54321 {
		t.Errorf("Expected TX123 to have LunchMoneyID 54321, got %d", syncedTx1.LunchMoneyID)
	}

	// Verify that TX456 was not changed
	syncedTx2, err := mockDB.GetTransactionByReference("TX456")
	if err != nil {
		t.Fatalf("Failed to get transaction TX456: %v", err)
	}
	if syncedTx2.LunchMoneyID != 12345 {
		t.Errorf("Expected TX456 to have LunchMoneyID 12345, got %d", syncedTx2.LunchMoneyID)
	}
}

func TestFilterUnsyncedTransactions(t *testing.T) {
	// Create mock database
	mockDB := db.NewMockDB()

	// Create mock LunchMoney client
	mockClient := &lm.MockLunchMoneyClient{
		Transactions: []models.Transaction{
			{
				ReferenceNumber: "TX456",
				Amount: models.Amount{
					Value:    "50.00",
					Currency: "USD",
				},
				Merchant: &models.Merchant{
					Name: "Test Merchant 2",
				},
				Date:         time.Now().Format(time.DateOnly),
				LunchMoneyID: 12345,
			},
			{
				ReferenceNumber: "TX789",
				Amount: models.Amount{
					Value:    "75.00",
					Currency: "USD",
				},
				Merchant: &models.Merchant{
					Name: "Test Merchant 3",
				},
				Date:         time.Now().Format(time.DateOnly),
				LunchMoneyID: 67890,
			},
		},
	}

	// Create the syncer
	syncer := &LunchMoneySyncer{
		client:   mockClient,
		database: mockDB,
	}

	// Create test transactions
	tx1 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 1",
				Address: &models.Address{},
			},
			Date:         time.Now().Format(time.DateOnly),
			LunchMoneyID: 0, // Not synced yet
		},
		SourceAccountName: "Test Account",
	}

	tx2 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX456",
			Amount: models.Amount{
				Value:    "50.00",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 2",
				Address: &models.Address{},
			},
			Date:         time.Now().Format(time.DateOnly),
			LunchMoneyID: 0, // Not synced yet, but matches an existing LunchMoney transaction
		},
		SourceAccountName: "Test Account",
	}

	tx3 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX789",
			Amount: models.Amount{
				Value:    "75.00",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 3",
				Address: &models.Address{},
			},
			Date:         time.Now().Format(time.DateOnly),
			LunchMoneyID: 67890, // Already synced
		},
		SourceAccountName: "Test Account",
	}

	transactions := []*models.TransactionWithAccount{tx1, tx2, tx3}

	// Test filtering unsynced transactions
	unsynced, needUpdate, err := syncer.filterUnsyncedTransactions(context.Background(), transactions)
	if err != nil {
		t.Fatalf("Failed to filter unsynced transactions: %v", err)
	}

	// Verify unsynced transactions
	if len(unsynced) != 1 {
		t.Errorf("Expected 1 unsynced transaction, got %d", len(unsynced))
	}
	if len(unsynced) > 0 && unsynced[0].ReferenceNumber != "TX123" {
		t.Errorf("Expected unsynced transaction TX123, got %s", unsynced[0].ReferenceNumber)
	}

	// Verify transactions needing update
	if len(needUpdate) != 1 {
		t.Errorf("Expected 1 transaction needing update, got %d", len(needUpdate))
	}
	if len(needUpdate) > 0 && needUpdate[0].ReferenceNumber != "TX456" {
		t.Errorf("Expected transaction needing update TX456, got %s", needUpdate[0].ReferenceNumber)
	}
	if len(needUpdate) > 0 && needUpdate[0].LunchMoneyID != 12345 {
		t.Errorf("Expected TX456 to have LunchMoneyID 12345, got %d", needUpdate[0].LunchMoneyID)
	}
}

func TestEnrichWithAccounts(t *testing.T) {
	// Create mock database
	mockDB := db.NewMockDB()
	// Create mock LunchMoney client
	mockClient := &lm.MockLunchMoneyClient{
		Accounts: []models.LunchMoneyAccount{
			{LunchMoneyId: 1, Name: "Test Account"},
		},
	}

	// Create a mock account selector
	mockSelector := NewAccountSelectorWithClient(mockClient, mockDB)
	mockSelector.selectedAccount = &models.AccountMapping{
		LunchMoneyId: 1,
		ExternalName: "Test Account",
	}

	// Create the syncer
	syncer := &LunchMoneySyncer{
		client:          mockClient,
		accountSelector: mockSelector,
	}

	// Create test transactions
	tx1 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX123",
			Amount: models.Amount{
				Value:    "25.99",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 1",
				Address: &models.Address{},
			},
			Date: time.Now().Format(time.DateOnly),
		},
		SourceAccountName: "Test Account",
	}

	tx2 := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: "TX456",
			Amount: models.Amount{
				Value:    "50.00",
				Currency: "USD",
			},
			Merchant: &models.Merchant{
				Name:    "Test Merchant 2",
				Address: &models.Address{},
			},
			Date: time.Now().Format(time.DateOnly),
		},
		SourceAccountName: "Test Account",
	}

	transactions := []*models.TransactionWithAccount{tx1, tx2}

	// Test enriching transactions with accounts
	enriched, err := syncer.enrichWithAccounts(context.Background(), transactions)
	if err != nil {
		t.Fatalf("Failed to enrich transactions: %v", err)
	}

	// Verify enriched transactions
	if len(enriched) != 2 {
		t.Errorf("Expected 2 enriched transactions, got %d", len(enriched))
	}

	for _, tx := range enriched {
		if tx.Mapping == nil {
			t.Errorf("Expected transaction to have an account, got nil")
		} else if tx.Mapping.LunchMoneyId != 1 {
			t.Errorf("Expected account ID 1, got %d", tx.Mapping.LunchMoneyId)
		} else if tx.Mapping.ExternalName != "Test Account" {
			t.Errorf("Expected account name 'Test Account', got '%s'", tx.Mapping.ExternalName)
		}
	}
}
