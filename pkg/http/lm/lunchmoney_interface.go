package lm

import (
	"context"

	"github.com/icco/lunchmoney"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// LunchMoneyClientInterface defines the interface for LunchMoney API operations
type LunchMoneyClientInterface interface {
	ListAccounts(ctx context.Context) ([]models.LunchMoneyAccount, error)
	ListTransaction(ctx context.Context, filter *lunchmoney.TransactionFilters) ([]models.Transaction, error)
	InsertTransactions(ctx context.Context, transactions []*models.TransactionWithAccountMapping) ([]int64, error)
}

// Ensure LunchMoneyClient implements LunchMoneyClientInterface
var _ LunchMoneyClientInterface = (*LunchMoneyClient)(nil)

// Ensure MockLunchMoneyClient implements LunchMoneyClientInterface
var _ LunchMoneyClientInterface = (*MockLunchMoneyClient)(nil)
