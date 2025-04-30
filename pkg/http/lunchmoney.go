package http

import (
	"context"
	"strconv"
	"strings"

	"github.com/vpineda1996/sandwich-sync/pkg/models"

	"github.com/icco/lunchmoney"
)

type LunchMoneyClient struct {
	client *lunchmoney.Client
}

func NewLunchMoneyClient(ctx context.Context, apiKey string) (*LunchMoneyClient, error) {
	client, err := lunchmoney.NewClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &LunchMoneyClient{
		client: client,
	}, nil
}

func (c *LunchMoneyClient) ListInstitutions(ctx context.Context) ([]models.Institution, error) {
	// Fetch institutions from the LunchMoney API
	assets, err := c.client.GetAssets(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the transaction's merchant name matches any institution's card name
	institutions := make([]models.Institution, 0)
	for _, asset := range assets {
		institutions = append(institutions, models.Institution{
			Id:   asset.ID,
			Name: asset.Name,
		})
	}

	return institutions, nil
}

func (c *LunchMoneyClient) ListTransaction(ctx context.Context, filter *lunchmoney.TransactionFilters) ([]models.Transaction, error) {
	lmTrns, err := c.client.GetTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	var translatedTrns []models.Transaction
	for _, lmTransaction := range lmTrns {

		translatedTrns = append(translatedTrns, models.Transaction{
			ReferenceNumber: lmTransaction.ExternalID,
			Merchant: &models.Merchant{
				Name:         lmTransaction.Payee,
				CategoryCode: strconv.FormatInt(lmTransaction.CategoryID, 10),
			},
			Amount: &models.Amount{
				Value:    lmTransaction.Amount,
				Currency: lmTransaction.Currency,
			},
			LunchMoneyID: lmTransaction.ID,
			Date:         lmTransaction.Date,
		})
	}
	return translatedTrns, nil
}

func (c *LunchMoneyClient) InsertTransactions(ctx context.Context, transactions []*models.TransactionWithInstitution) ([]int64, error) {
	var lmTrns []lunchmoney.InsertTransaction
	for _, transaction := range transactions {
		// Create a new transaction object for LunchMoney
		lmTrns = append(lmTrns, lunchmoney.InsertTransaction{
			Date:       transaction.Date,
			Amount:     transaction.Amount.Value,
			Currency:   strings.ToLower(transaction.Amount.Currency),
			ExternalID: transaction.ReferenceNumber,
			Payee:      transaction.Merchant.Name,
			AssetID:    &transaction.Institution.Id,
		})
	}

	// Insert the transaction into LunchMoney
	response, err := c.client.InsertTransactions(ctx, lunchmoney.InsertTransactionsRequest{
		ApplyRules:   true,
		Transactions: lmTrns,
	})

	if err != nil {
		return nil, err
	}

	return response.IDs, nil
}
