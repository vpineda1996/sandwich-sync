package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"

	"github.com/icco/lunchmoney"
	"github.com/samber/lo"
)

func (l *LunchMoneySyncer) SyncTransactions(ctx context.Context) error {
	// fetch the transactions from our local database
	transactions, err := l.database.GetTransactions()
	if err != nil {
		return err
	}
	log.Info().Int("count", len(transactions)).
		Msg("Fetched transactions from local database")

	// only consider transactions that are less than 30 days old
	recentTransactions := make([]*models.TransactionWithAccount, 0)
	for _, transaction := range transactions {
		transactionDate, err := time.Parse(time.DateOnly, transaction.Date)
		if err != nil {
			return err
		}
		if transactionDate.After(time.Now().Add(-30 * 24 * time.Hour)) {
			recentTransactions = append(recentTransactions, transaction)
		}
	}

	log.Info().Int("count", len(recentTransactions)).Msg("Filtered recent transactions")

	// filter transactions to only those that are not already synced
	unsyncedTransactions, syncedNeededToUpdateTransactions, err := l.filterUnsyncedTransactions(ctx, recentTransactions)
	if err != nil {
		return err
	}

	if len(unsyncedTransactions) != 0 {
		enrichUnsyncedTransactions, err := l.enrichWithAccounts(ctx, unsyncedTransactions)
		if err != nil {
			return err
		}

		// filter out transactions that are marked for no sync
		unsyncedTransactions, enrichUnsyncedTransactions, err = l.
			filterOutNoSyncTransactions(unsyncedTransactions, enrichUnsyncedTransactions)
		if err != nil {
			return err
		}

		if len(unsyncedTransactions) != 0 {
			// sync the transactions with the LunchMoney API
			log.Info().Int("count", len(unsyncedTransactions)).Msg("Unsynced transactions with LunchMoney")
			insertionIds, err := l.client.InsertTransactions(ctx, enrichUnsyncedTransactions)
			if err != nil {
				return err
			}

			if len(insertionIds) != len(enrichUnsyncedTransactions) {
				return fmt.Errorf("failed to insert all transactions, expected %d, got %d", len(enrichUnsyncedTransactions), len(insertionIds))
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

func (l *LunchMoneySyncer) filterOutNoSyncTransactions(unsyncedTransactions []*models.TransactionWithAccount,
	enrichUnsyncedTransactions []*models.TransactionWithAccountMapping) ([]*models.TransactionWithAccount, []*models.TransactionWithAccountMapping, error) {
	ut, eut := make([]*models.TransactionWithAccount, 0), make([]*models.TransactionWithAccountMapping, 0)
	for i, e := range enrichUnsyncedTransactions {
		shouldSync, err := l.database.IsSyncOptionEnabled(e.Mapping.LunchMoneyId, models.SyncOptionTransactions)
		if err != nil {
			return nil, nil, err
		}

		if shouldSync {
			ut = append(ut, unsyncedTransactions[i])
			eut = append(eut, e)
		}
	}
	return ut, eut, nil
}

func (l *LunchMoneySyncer) enrichWithAccounts(ctx context.Context,
	unsyncedTransactions []*models.TransactionWithAccount) ([]*models.TransactionWithAccountMapping, error) {
	enrichUnsyncedTransactions := make([]*models.TransactionWithAccountMapping, 0)

	for _, transaction := range unsyncedTransactions {
		// find the account for the transaction
		account, err := l.accountMapper.FindPossibleAccountForTransaction(ctx, transaction)
		if err != nil {
			return nil, err
		}

		if account == nil {
			log.Info().Str("transactionId", transaction.ReferenceNumber).Msg("No account found for transaction")
			continue
		}
		// create a new transaction with the account
		transactionWithAccount := &models.TransactionWithAccountMapping{
			Transaction: transaction.Transaction,
			Mapping:     account,
		}
		enrichUnsyncedTransactions = append(enrichUnsyncedTransactions, transactionWithAccount)
	}
	return enrichUnsyncedTransactions, nil
}

func (l *LunchMoneySyncer) filterUnsyncedTransactions(ctx context.Context,
	transactions []*models.TransactionWithAccount) ([]*models.TransactionWithAccount, []*models.TransactionWithAccount, error) {
	missingLunchId := make([]*models.TransactionWithAccount, 0)
	for _, transaction := range transactions {
		if transaction.ReferenceNumber == "" {
			// Skip transactions without a reference number
			continue
		}

		if transaction.LunchMoneyID < 0 {
			// Skip transactions with a negative LunchMoneyID
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
	missingUpdate := make([]*models.TransactionWithAccount, 0)
	unsynced := make([]*models.TransactionWithAccount, 0)
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
				log.Info().Str("transactionId", transaction.ReferenceNumber).
					Int64("lunchId", lunchTransaction.LunchMoneyID).Msg("Transaction is already synced with LunchMoney")
				transaction.LunchMoneyID = lunchTransaction.LunchMoneyID
				missingUpdate = append(missingUpdate, transaction)
				transactionSynced = true
				break
			}
		}

		if transactionSynced {
			continue
		}

		// This transaction is not synced
		unsynced = append(unsynced, transaction)
	}

	return unsynced, missingUpdate, nil
}
