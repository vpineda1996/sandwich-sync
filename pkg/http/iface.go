package http

import (
	"context"

	"github.com/vpineda1996/sandwich-sync/pkg/http/rogers"
	"github.com/vpineda1996/sandwich-sync/pkg/http/ws"
	"github.com/vpineda1996/sandwich-sync/pkg/models"
)

type TransactionFetcher interface {
	FetchTransactions(ctx context.Context) ([]models.Transaction, error)
}

var (
	_ TransactionFetcher = &rogers.RogersBankClient{}
	_ TransactionFetcher = &ws.WealthsimpleClient{}
)
