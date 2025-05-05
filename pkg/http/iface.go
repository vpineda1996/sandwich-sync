package http

import (
	"context"

	"github.com/vpnda/sandwich-sync/pkg/http/rogers"
	"github.com/vpnda/sandwich-sync/pkg/http/ws"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

type TransactionFetcher interface {
	FetchTransactions(ctx context.Context) ([]models.Transaction, error)
}

var (
	_ TransactionFetcher = &rogers.RogersBankClient{}
	_ TransactionFetcher = &ws.WealthsimpleClient{}
)
