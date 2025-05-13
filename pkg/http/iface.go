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

type BalanceFetcher interface {
	FetchAccountBalances(ctx context.Context) ([]models.ExternalAccount, error)
}
