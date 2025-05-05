package services

import (
	"context"
	"fmt"
	"time"

	"github.com/vpnda/sandwich-sync/db"
	"github.com/vpnda/sandwich-sync/pkg/http/lm"
	"github.com/vpnda/sandwich-sync/pkg/models"

	"github.com/icco/lunchmoney"
	"github.com/samber/lo"
)

type LunchMoneySyncer struct {
	client          lm.LunchMoneyClientInterface
	database        db.DBInterface
	accountSelector *InstitutionSelector
	forceSync       bool
}

func NewLunchMoneySyncer(ctx context.Context, apiKey string, database db.DBInterface) (*LunchMoneySyncer, error) {
	c, err := lm.NewLunchMoneyClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	as, err := NewInstitutionSelector(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return &LunchMoneySyncer{
		client:          c,
		database:        database,
		accountSelector: as,
	}, nil
}

func (l *LunchMoneySyncer) SyncTransactions(ctx context.Context) error {
	// fetch the transactions from our local database
	transactions, err := l.database.GetTransactions()
	if err != nil {
		return err
	}
	fmt.Printf("Fetched %d transactions from local database\n", len(transactions))

	// only consider transactions that are less than 30 days old
	recentTransactions := make([]*models.Transaction, 0)
	for _, transaction := range transactions {
		transactionDate, err := time.Parse(time.DateOnly, transaction.Date)
		if err != nil {
			return err
		}
		if transactionDate.After(time.Now().Add(-30 * 24 * time.Hour)) {
			recentTransactions = append(recentTransactions, transaction)
		}
	}

	fmt.Printf("Filtered to %d recent transactions\n", len(recentTransactions))

	// filter transactions to only those that are not already synced
	unsyncedTransactions, syncedNeededToUpdateTransactions, err := l.filterUnsyncedTransactions(ctx, recentTransactions)
	if err != nil {
		return err
	}

	if len(unsyncedTransactions) != 0 {
		// sync the transactions with the LunchMoney API
		fmt.Printf("Syncing %d unsynced transactions with LunchMoney\n", len(unsyncedTransactions))
		enrichUnsyncedTransactions, err := l.enrichWithInstitutions(ctx, unsyncedTransactions)
		if err != nil {
			return err
		}

		insertionIds, err := l.client.InsertTransactions(ctx, enrichUnsyncedTransactions)
		if err != nil {
			return err
		}

		// update the local database with the insertion IDs
		for i, transaction := range unsyncedTransactions {
			if i < len(insertionIds) {
				transaction.LunchMoneyID = insertionIds[i]
				if err := l.database.UpdateTransaction(transaction); err != nil {
					return err
				}
			}
		}
	}

	if len(syncedNeededToUpdateTransactions) != 0 {
		// update the transactions locally
		for _, transaction := range syncedNeededToUpdateTransactions {
			// update the transaction in the local database
			if err := l.database.UpdateTransaction(transaction); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *LunchMoneySyncer) enrichWithInstitutions(ctx context.Context,
	unsyncedTransactions []*models.Transaction) ([]*models.TransactionWithInstitution, error) {
	enrichUnsyncedTransactions := make([]*models.TransactionWithInstitution, 0)

	for _, transaction := range unsyncedTransactions {
		// find the institution for the transaction
		institution, err := l.accountSelector.FindPossibleInstitutionForTransaction(ctx, transaction)
		if err != nil {
			return nil, err
		}

		if institution == nil {
			fmt.Printf("No institution found for transaction %s\n", transaction.ReferenceNumber)
			continue
		}
		// create a new transaction with the institution
		transactionWithInstitution := &models.TransactionWithInstitution{
			Transaction: *transaction,
			Institution: institution,
		}
		enrichUnsyncedTransactions = append(enrichUnsyncedTransactions, transactionWithInstitution)
	}
	return enrichUnsyncedTransactions, nil
}

func (l *LunchMoneySyncer) filterUnsyncedTransactions(ctx context.Context, transactions []*models.Transaction) ([]*models.Transaction, []*models.Transaction, error) {
	missingLunchId := make([]*models.Transaction, 0)
	for _, transaction := range transactions {
		if transaction.ReferenceNumber == "" {
			// Skip transactions without a reference number
			continue
		}

		if l.forceSync || transaction.LunchMoneyID == 0 {
			missingLunchId = append(missingLunchId, transaction)
		}
	}

	// Check if there are any unsynced transactions by getting transactions and match on
	// date, amount and merchant name
	lunchTransactions, err := l.client.ListTransaction(ctx, &lunchmoney.TransactionFilters{
		StartDate: lo.ToPtr(time.Now().Add(-30 * 24 * time.Hour).Format(time.DateOnly)),
		EndDate:   lo.ToPtr(time.Now().Format(time.DateOnly)),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch transactions from LunchMoney: %w", err)
	}

	// Filter out transactions that are already synced
	missingUpdate := make([]*models.Transaction, 0)
	unsynced := make([]*models.Transaction, 0)
	for _, transaction := range missingLunchId {
		transactionSynced := false
		for _, lunchTransaction := range lunchTransactions {
			amountEq, err := transaction.Amount.ToMoney().Equals(lunchTransaction.Amount.ToMoney())
			if err != nil {
				continue
			}
			if (transaction.Date == lunchTransaction.Date &&
				amountEq &&
				transaction.Merchant.Name == lunchTransaction.Merchant.Name) || transaction.ReferenceNumber == lunchTransaction.ReferenceNumber {
				// This transaction is already synced
				fmt.Printf("Transaction %s is already synced with LunchMoney (id: %d) \n", transaction.ReferenceNumber, lunchTransaction.LunchMoneyID)
				transaction.LunchMoneyID = lunchTransaction.LunchMoneyID
				missingUpdate = append(missingUpdate, transaction)
				transactionSynced = true
				break
			}
		}

		if !transactionSynced {
			// This transaction is not synced
			fmt.Printf("Transaction [%s] from %s is not synced with LunchMoney \n", transaction.ReferenceNumber, transaction.Merchant.Name)
			unsynced = append(unsynced, transaction)
		}
	}

	return unsynced, missingUpdate, nil
}
