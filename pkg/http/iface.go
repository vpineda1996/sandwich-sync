package http

import (
	"context"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

type Fetcher interface {
	TransactionFetcher
	BalanceFetcher
}

type TransactionFetcher interface {
	FetchTransactions(ctx context.Context) ([]models.TransactionWithAccount, error)
}

type BalanceStorer interface {
	UpsertAccountBalance(externalAccountName string, balance models.Amount) error
}

type BalanceFetcher interface {
	UpdateAccountBalances(ctx context.Context, balanceStorage BalanceStorer) error
}
