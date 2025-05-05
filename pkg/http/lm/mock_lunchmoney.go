package lm

import (
	"context"

	"github.com/icco/lunchmoney"
	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

// MockLunchMoneyClient is a mock implementation of the LunchMoneyClient for testing
type MockLunchMoneyClient struct {
	// Mock data to return
	Institutions []models.Institution
	Transactions []models.Transaction
	InsertedIDs  []int64

	// Error values to return
	ListInstitutionsErr   error
	ListTransactionErr    error
	InsertTransactionsErr error
}

// NewMockLunchMoneyClient creates a new mock LunchMoney client
func NewMockLunchMoneyClient() *MockLunchMoneyClient {
	return &MockLunchMoneyClient{
		Institutions: []models.Institution{},
		Transactions: []models.Transaction{},
		InsertedIDs:  []int64{},
	}
}

// ListInstitutions returns the mock institutions
func (m *MockLunchMoneyClient) ListInstitutions(ctx context.Context) ([]models.Institution, error) {
	if m.ListInstitutionsErr != nil {
		return nil, m.ListInstitutionsErr
	}
	return m.Institutions, nil
}

// ListTransaction returns the mock transactions
func (m *MockLunchMoneyClient) ListTransaction(ctx context.Context, filter *lunchmoney.TransactionFilters) ([]models.Transaction, error) {
	if m.ListTransactionErr != nil {
		return nil, m.ListTransactionErr
	}
	return m.Transactions, nil
}

// InsertTransactions returns the mock inserted IDs
func (m *MockLunchMoneyClient) InsertTransactions(ctx context.Context, transactions []*models.TransactionWithInstitution) ([]int64, error) {
	if m.InsertTransactionsErr != nil {
		return nil, m.InsertTransactionsErr
	}
	return m.InsertedIDs, nil
}
