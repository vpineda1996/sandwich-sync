package lm

import (
	"context"
	"time"

	"github.com/icco/lunchmoney"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// MockLunchMoneyClient is a mock implementation of the LunchMoneyClient for testing
type MockLunchMoneyClient struct {
	// Mock data to return
	Accounts     []models.LunchMoneyAccount
	Transactions []models.Transaction
	InsertedIDs  []int64

	// Error values to return
	ListAccountsErr       error
	ListTransactionErr    error
	InsertTransactionsErr error
}

// UpdateAccountBalance implements LunchMoneyClientInterface.
func (m *MockLunchMoneyClient) UpdateAccountBalance(ctx context.Context, id int64, balance models.Amount, since *time.Time) error {
	panic("unimplemented")
}

// NewMockLunchMoneyClient creates a new mock LunchMoney client
func NewMockLunchMoneyClient() *MockLunchMoneyClient {
	return &MockLunchMoneyClient{
		Accounts:     []models.LunchMoneyAccount{},
		Transactions: []models.Transaction{},
		InsertedIDs:  []int64{},
	}
}

// ListAccounts returns the mock accounts
func (m *MockLunchMoneyClient) ListAccounts(ctx context.Context) ([]models.LunchMoneyAccount, error) {
	if m.ListAccountsErr != nil {
		return nil, m.ListAccountsErr
	}
	return m.Accounts, nil
}

// ListTransaction returns the mock transactions
func (m *MockLunchMoneyClient) ListTransaction(ctx context.Context, filter *lunchmoney.TransactionFilters) ([]models.Transaction, error) {
	if m.ListTransactionErr != nil {
		return nil, m.ListTransactionErr
	}
	return m.Transactions, nil
}

// InsertTransactions returns the mock inserted IDs
func (m *MockLunchMoneyClient) InsertTransactions(ctx context.Context, transactions []*models.TransactionWithAccountMapping) ([]int64, error) {
	if m.InsertTransactionsErr != nil {
		return nil, m.InsertTransactionsErr
	}
	return m.InsertedIDs, nil
}
