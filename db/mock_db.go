package db

import (
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

// MockDB is a mock implementation of the DB for testing
type MockDB struct {
	// Mock data storage
	Transactions map[string]*models.TransactionWithAccount
	// Mock data for account mappings
	AccountMappings map[string]*models.AccountMapping

	// Error values to return
	GetTransactionsErr           error
	GetTransactionByReferenceErr error
	SaveTransactionErr           error
	UpdateTransactionErr         error
	RemoveTransactionErr         error
	AddManualTransactionErr      error

	GetAccountMappingErr    error
	UpsertAccountMappingErr error
}

// GetAccounts implements DBInterface.
func (m *MockDB) GetAccounts() ([]models.LunchMoneyAccount, error) {
	panic("unimplemented")
}

// UpsertAccountBalance implements DBInterface.
func (m *MockDB) UpsertAccountBalance(lunchMoneyId string, balance models.Amount) error {
	panic("unimplemented")
}

// GetAccountMapping implements DBInterface.
func (m *MockDB) GetAccountMapping(externalId string) (*models.AccountMapping, error) {
	if m.GetAccountMappingErr != nil {
		return nil, m.GetAccountMappingErr
	}

	if am, ok := m.AccountMappings[externalId]; ok {
		return am, nil
	}

	return nil, nil
}

// UpsertAccountMapping implements DBInterface.
func (m *MockDB) UpsertAccountMapping(am *models.AccountMapping) error {
	if m.UpsertAccountMappingErr != nil {
		return m.UpsertAccountMappingErr
	}

	if m.AccountMappings == nil {
		m.AccountMappings = make(map[string]*models.AccountMapping)
	}

	m.AccountMappings[am.ExternalName] = am
	return nil
}

// NewMockDB creates a new mock database
func NewMockDB() *MockDB {
	return &MockDB{
		Transactions: make(map[string]*models.TransactionWithAccount),
	}
}

// GetTransactions returns all transactions in the mock database
func (m *MockDB) GetTransactions() ([]*models.TransactionWithAccount, error) {
	if m.GetTransactionsErr != nil {
		return nil, m.GetTransactionsErr
	}

	transactions := make([]*models.TransactionWithAccount, 0, len(m.Transactions))
	for _, tx := range m.Transactions {
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetTransactionByReference returns a transaction by its reference number
func (m *MockDB) GetTransactionByReference(referenceNumber string) (*models.TransactionWithAccount, error) {
	if m.GetTransactionByReferenceErr != nil {
		return nil, m.GetTransactionByReferenceErr
	}

	tx, ok := m.Transactions[referenceNumber]
	if !ok {
		return nil, nil
	}

	return tx, nil
}

// SaveTransaction saves a transaction to the mock database
func (m *MockDB) SaveTransaction(tx *models.TransactionWithAccount) error {
	if m.SaveTransactionErr != nil {
		return m.SaveTransactionErr
	}

	m.Transactions[tx.ReferenceNumber] = tx
	return nil
}

// UpdateTransaction updates a transaction in the mock database
func (m *MockDB) UpdateTransaction(tx *models.TransactionWithAccount) error {
	if m.UpdateTransactionErr != nil {
		return m.UpdateTransactionErr
	}

	if _, ok := m.Transactions[tx.ReferenceNumber]; !ok {
		return fmt.Errorf("no transaction found with reference number: %s", tx.ReferenceNumber)
	}

	m.Transactions[tx.ReferenceNumber] = tx
	return nil
}

// RemoveTransaction removes a transaction from the mock database
func (m *MockDB) RemoveTransaction(referenceNumber string) error {
	if m.RemoveTransactionErr != nil {
		return m.RemoveTransactionErr
	}

	if _, ok := m.Transactions[referenceNumber]; !ok {
		return fmt.Errorf("no transaction found with reference number: %s", referenceNumber)
	}

	delete(m.Transactions, referenceNumber)
	return nil
}

// AddManualTransaction adds a manually created transaction to the mock database
func (m *MockDB) AddManualTransaction(tx *models.TransactionWithAccount) error {
	if m.AddManualTransactionErr != nil {
		return m.AddManualTransactionErr
	}

	if _, ok := m.Transactions[tx.ReferenceNumber]; ok {
		return fmt.Errorf("transaction with reference number %s already exists", tx.ReferenceNumber)
	}

	m.Transactions[tx.ReferenceNumber] = tx
	return nil
}

// Initialize is a no-op for the mock database
func (m *MockDB) Initialize() error {
	return nil
}

// Close is a no-op for the mock database
func (m *MockDB) Close() error {
	return nil
}
