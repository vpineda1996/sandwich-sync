package ws

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"
	"github.com/vpnda/sandwich-sync/pkg/utils"
	"github.com/vpnda/wsfetch/pkg/client/generated"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (w *WealthsimpleClient) FetchAccountBalances(ctx context.Context) ([]models.ExternalAccount, error) {
	accounts, err := w.c.GetAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	var externalAccounts []models.ExternalAccount
	for _, account := range accounts {
		if account.ClosedAt != nil {
			log.Info().Str("accountId", account.Id).Msgf("Skipping retrieving closed account balance")
			continue
		}

		balance := models.Amount{
			Value:    account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Amount,
			Currency: account.GetFinancials().CurrentCombined.NetLiquidationValueV2.Currency,
		}
		log.Info().Msgf("Found balance for account %s: %s", account.Id, balance.ToMoney().Display())
		externalAccounts = append(externalAccounts, models.ExternalAccount{
			Name:        account.Id,
			Balance:     balance,
			Description: createHumanFriendsDescriptionForAccount(account),
		})
	}

	return externalAccounts, nil
}

func createHumanFriendsDescriptionForAccount(account generated.AccountWithFinancials) string {
	var accountType string
	if account.UnifiedAccountType != nil {
		accountType = strings.ReplaceAll(*account.UnifiedAccountType, "_", " ")
		accountType = utils.Capitalize(accountType)
	}

	if accountType != "" && account.Nickname != nil {
		return fmt.Sprintf("%s (%s)", *account.Nickname, accountType)
	}

	if accountType != "" && account.Currency != nil {
		return fmt.Sprintf("%s (%s)", accountType, *account.Currency)
	}

	return account.Id
}
