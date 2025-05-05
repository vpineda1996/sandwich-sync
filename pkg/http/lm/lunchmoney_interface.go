package lm

import (
	"context"

	"github.com/icco/lunchmoney"
	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

// LunchMoneyClientInterface defines the interface for LunchMoney API operations
type LunchMoneyClientInterface interface {
	ListInstitutions(ctx context.Context) ([]models.Institution, error)
	ListTransaction(ctx context.Context, filter *lunchmoney.TransactionFilters) ([]models.Transaction, error)
	InsertTransactions(ctx context.Context, transactions []*models.TransactionWithInstitution) ([]int64, error)
}

// Ensure LunchMoneyClient implements LunchMoneyClientInterface
var _ LunchMoneyClientInterface = (*LunchMoneyClient)(nil)

// Ensure MockLunchMoneyClient implements LunchMoneyClientInterface
var _ LunchMoneyClientInterface = (*MockLunchMoneyClient)(nil)
