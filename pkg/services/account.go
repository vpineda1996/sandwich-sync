package services

import (
	"context"
	"fmt"

	"github.com/vpnda/sandwich-sync/db"
	"github.com/vpnda/sandwich-sync/pkg/http/lm"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

const (
	DefaultAccountName = "Default Account"
)

type AccountSelector struct {
	client          lm.LunchMoneyClientInterface
	db              db.DBInterface
	selectedAccount *models.AccountMapping
}

func NewAccountSelector(ctx context.Context, apiKey string, database db.DBInterface) (*AccountSelector, error) {
	c, err := lm.NewLunchMoneyClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return &AccountSelector{
		client: c,
		db:     database,
	}, nil
}

// NewAccountSelectorWithClient creates a new account selector with a provided client
func NewAccountSelectorWithClient(client lm.LunchMoneyClientInterface, database db.DBInterface) *AccountSelector {
	return &AccountSelector{
		client: client,
		db:     database,
	}
}

func (is *AccountSelector) FindPossibleAccountForTransaction(ctx context.Context, transaction *models.TransactionWithAccount) (*models.AccountMapping, error) {
	// Fetch mapping from the database
	mapping, err := is.db.GetAccountMapping(transaction.ReferenceNumber)
	if err != nil {
		return nil, err
	}

	if mapping != nil {
		return mapping, nil
	}

	if is.selectedAccount != nil {
		return is.selectedAccount, nil
	}

	fmt.Printf("Could not find account for transaction [%s] %s (%s). Please select one:\n",
		transaction.ReferenceNumber, transaction.Merchant.Name, transaction.Amount.ToMoney().Display())
	return is.selectAccountInteractive(transaction.SourceAccountName)
}

func (is *AccountSelector) selectAccountInteractive(sourceAccountName string) (*models.AccountMapping, error) {
	accounts, err := is.client.ListAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	for i, account := range accounts {
		fmt.Printf("\t %-2d. %s\n", i, account.Name)
	}

	var selection int
	fmt.Printf("Enter the number of the account you want to map %q: ", sourceAccountName)
	_, err = fmt.Scan(&selection)
	if err != nil || selection < 0 || selection >= len(accounts) {
		return nil, fmt.Errorf("invalid selection")
	}

	selected := &accounts[selection]
	mapping := &models.AccountMapping{
		LunchMoneyId: selected.LunchMoneyId,
		ExternalName: sourceAccountName,
		IsPlaid:      selected.IsPlaid,
	}

	// Save the mapping to the database
	if sourceAccountName != DefaultAccountName {
		err = is.db.UpsertAccountMapping(mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to save account mapping: %w", err)
		}
	}

	return mapping, nil
}

func (is *AccountSelector) SelectDefaultAccount() error {
	sa, err := is.selectAccountInteractive(DefaultAccountName)
	if err != nil {
		return err
	}
	is.selectedAccount = sa
	return nil
}
