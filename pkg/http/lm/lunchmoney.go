package lm

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/vpnda/sandwich-sync/pkg/models"

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

func (c *LunchMoneyClient) ListAccounts(ctx context.Context) ([]models.LunchMoneyAccount, error) {
	// Fetch accounts from the LunchMoney API
	assets, err := c.client.GetAssets(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the transaction's merchant name matches any account's card name
	accounts := make([]models.LunchMoneyAccount, 0)
	for _, asset := range assets {
		accounts = append(accounts, models.LunchMoneyAccount{
			LunchMoneyId: asset.ID,
			Name:         asset.Name,
			DisplayName:  asset.DisplayName,
			Balance: models.Amount{
				Value:    asset.Balance,
				Currency: asset.Currency,
			},
			BalanceLastUpdated: &asset.BalanceAsOf,
		})
	}

	return accounts, nil
}

// UpdateAccountBalance implements LunchMoneyClientInterface.
func (c *LunchMoneyClient) UpdateAccountBalance(ctx context.Context, id int64, balance models.Amount, since *time.Time) error {
	// Update the account balance in LunchMoney
	_, err := c.client.UpdateAsset(ctx, id, &lunchmoney.UpdateAsset{
		Balance:     &balance.Value,
		Currency:    &balance.Currency,
		BalanceAsOf: lo.ToPtr(since.Format(time.RFC3339)),
	})
	if err != nil {
		return err
	}

	return nil
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
			Amount: models.Amount{
				Value:    lmTransaction.Amount,
				Currency: lmTransaction.Currency,
			},
			LunchMoneyID: lmTransaction.ID,
			Date:         lmTransaction.Date,
		})
	}
	return translatedTrns, nil
}

func (c *LunchMoneyClient) InsertTransactions(ctx context.Context, transactions []*models.TransactionWithAccountMapping) ([]int64, error) {
	var lmTrns []lunchmoney.InsertTransaction
	for _, transaction := range transactions {
		// Create a new transaction object for LunchMoney
		lmTrns = append(lmTrns, lunchmoney.InsertTransaction{
			Date:       transaction.Date,
			Amount:     transaction.Amount.Value,
			Currency:   strings.ToLower(transaction.Amount.Currency),
			ExternalID: transaction.ReferenceNumber,
			Payee:      transaction.Merchant.Name,
			AssetID:    &transaction.Mapping.LunchMoneyId,
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
