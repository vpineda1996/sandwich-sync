package db

import (
	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

// DBInterface defines the interface for database operations
type DBInterface interface {
	Initialize() error
	Close() error
	GetTransactions() ([]*models.Transaction, error)
	GetTransactionByReference(referenceNumber string) (*models.Transaction, error)
	SaveTransaction(tx *models.Transaction) error
	UpdateTransaction(tx *models.Transaction) error
	RemoveTransaction(referenceNumber string) error
	AddManualTransaction(tx *models.Transaction) error
}

// Ensure DB implements DBInterface
var _ DBInterface = (*DB)(nil)

// Ensure MockDB implements DBInterface
var _ DBInterface = (*MockDB)(nil)
