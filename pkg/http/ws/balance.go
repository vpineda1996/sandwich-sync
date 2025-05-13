package ws

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (w *WealthsimpleClient) FetchAccountBalances(ctx context.Context) ([]models.ExternalAccount, error) {
	accounts, err := w.c.GetAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var externalAccounts []models.ExternalAccount
	for _, account := range accounts {
		balance := models.Amount{
			Value:    account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Amount,
			Currency: account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Currency,
		}
		log.Info().Msgf("Found balance for account %s: %s", account.Id, balance.ToMoney().Display())
		externalAccounts = append(externalAccounts, models.ExternalAccount{
			Name:    account.Id,
			Balance: balance,
		})
	}

	return externalAccounts, nil
}
