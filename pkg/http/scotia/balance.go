package scotia

import (
	"context"
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/http"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (s *ScotiaClient) UpdateAccountBalances(ctx context.Context, balanceStorage http.BalanceStorer) error {
	resp, r, err := s.apiClient.DefaultAPI.ApiAccountsSummaryGet(ctx).Execute()
	if err != nil {
		return fmt.Errorf("failed to fetch accounts summary: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("no response from API")
	}
	defer r.Body.Close()

	for _, account := range resp.Data.GetProducts() {
		externalAccountName := account.GetKey()
		primaryBalances := account.GetPrimaryBalances()
		if len(primaryBalances) == 0 {
			return fmt.Errorf("no primary balances found for account %s", externalAccountName)
		}
		balance := primaryBalances[0]
		amount := models.Amount{
			Value:    fmt.Sprintf("%.2f", balance.GetAmount()),
			Currency: balance.GetCurrencyCode(),
		}
		err = balanceStorage.UpsertAccountBalance(externalAccountName, amount)
		if err != nil {
			return fmt.Errorf("failed to upsert account balance: %w", err)
		}
	}
	return nil
}
