package services

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (l *LunchMoneySyncer) SyncBalances(ctx context.Context) error {
	//knownAccounts, err := l.database.GetAccounts()
	localAccounts, err := l.database.GetAccounts()
	if err != nil {
		return err
	}

	lunchMoneyAccounts, err := l.client.ListAccounts(ctx)
	if err != nil {
		return err
	}

	lunchMoneyMap := make(map[int64]models.LunchMoneyAccount)
	for _, account := range lunchMoneyAccounts {
		lunchMoneyMap[account.LunchMoneyId] = account
	}

	for _, localAccount := range localAccounts {
		_, ok := lunchMoneyMap[localAccount.LunchMoneyId]
		if !ok {
			// Account not found in LunchMoney, ignore it but warn user
			log.Warn().Str("account", localAccount.Name).Msg("Account not found in LunchMoney, not syncing balance")
			continue
		}
		lunchMoneyAccount := lunchMoneyMap[localAccount.LunchMoneyId]

		if localAccount.Balance.Value == lunchMoneyAccount.Balance.Value {
			continue
		}

		if localAccount.BalanceLastUpdated == nil || lunchMoneyAccount.BalanceLastUpdated == nil {
			log.Warn().Int64("account", localAccount.LunchMoneyId).Msg("Balance is out of sync, but we don't have a last updated date")
			continue
		}

		if localAccount.SyncStrategy&models.SyncOptionBalance == 0 {
			continue
		}

		if localAccount.BalanceLastUpdated.After(*lunchMoneyAccount.BalanceLastUpdated) {
			log.Info().Int64("account", localAccount.LunchMoneyId).Msg("Updating balance in LunchMoney to match local balance")
			// lower-case the curreny code to match LunchMoney's format
			localAccount.Balance.Currency = strings.ToLower(localAccount.Balance.Currency)
			err := l.client.UpdateAccountBalance(ctx,
				localAccount.LunchMoneyId,
				localAccount.Balance,
				localAccount.BalanceLastUpdated)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
