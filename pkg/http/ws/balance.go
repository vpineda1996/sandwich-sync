package ws

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/http"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (w *WealthsimpleClient) UpdateAccountBalances(ctx context.Context, balanceStorage http.BalanceStorer) error {
	accounts, err := w.c.GetAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get accounts: %w", err)
	}

	for _, account := range accounts {
		balance := models.Amount{
			Value:    account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Amount,
			Currency: account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Currency,
		}
		log.Info().Msgf("Found balance for account %s: %s", account.Id, balance.ToMoney().Display())
		err = balanceStorage.UpsertAccountBalance(account.Id, balance)
		if err != nil {
			return fmt.Errorf("failed to upsert account balance: %w", err)
		}
	}

	return nil
}
