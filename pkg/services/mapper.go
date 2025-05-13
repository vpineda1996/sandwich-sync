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

type AccountMapper struct {
	client          lm.LunchMoneyClientInterface
	db              db.DBInterface
	selectedAccount *models.AccountMapping
}

func NewAccountMapper(ctx context.Context, apiKey string, database db.DBInterface) (*AccountMapper, error) {
	c, err := lm.NewLunchMoneyClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return &AccountMapper{
		client: c,
		db:     database,
	}, nil
}

// NewAccountMapperWithClient creates a new account mapper with a provided client
func NewAccountMapperWithClient(client lm.LunchMoneyClientInterface, database db.DBInterface) *AccountMapper {
	return &AccountMapper{
		client: client,
		db:     database,
	}
}

func (is *AccountMapper) FindPossibleAccountForTransaction(ctx context.Context, transaction *models.TransactionWithAccount) (*models.AccountMapping, error) {
	// Fetch mapping from the database
	mapping, err := is.db.GetAccountMapping(transaction.SourceAccountName)
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

func (is *AccountMapper) FindPossibleAccountForExternal(ctx context.Context, externalAccount *models.ExternalAccount) (*models.AccountMapping, error) {
	// Fetch mapping from the database
	mapping, err := is.db.GetAccountMapping(externalAccount.Name)
	if err != nil {
		return nil, err
	}

	if mapping != nil {
		return mapping, nil
	}

	fmt.Printf("Could not find account for external account [%s] %s (%s). Please select one:\n",
		externalAccount.Name, externalAccount.Name, externalAccount.Balance.ToMoney().Display())
	return is.selectAccountInteractive(externalAccount.Name)
}

func (is *AccountMapper) selectAccountInteractive(sourceAccountName string) (*models.AccountMapping, error) {
	accounts, err := is.client.ListAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	for i, account := range accounts {
		fmt.Printf("\t %-2d. %s\n", i, account.Name)
	}

	var selection int
	fmt.Printf("Enter the number of the account you want to map %q or -1 to always ignore it: ", sourceAccountName)
	_, err = fmt.Scan(&selection)
	if err != nil || selection < -1 || selection >= len(accounts) {
		return nil, fmt.Errorf("invalid selection")
	}
	var mapping *models.AccountMapping
	if selection == -1 {
		mapping = &models.AccountMapping{
			LunchMoneyId: -1,
			ExternalName: sourceAccountName,
		}
	} else {
		mapping = &models.AccountMapping{
			LunchMoneyId: accounts[selection].LunchMoneyId,
			ExternalName: sourceAccountName,
			IsPlaid:      accounts[selection].IsPlaid,
		}
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

func (is *AccountMapper) SelectDefaultAccount() error {
	sa, err := is.selectAccountInteractive(DefaultAccountName)
	if err != nil {
		return err
	}
	is.selectedAccount = sa
	return nil
}
