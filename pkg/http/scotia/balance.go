package scotia

import (
	"context"
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (s *ScotiaClient) FetchAccountBalances(ctx context.Context) ([]models.ExternalAccount, error) {
	resp, r, err := s.apiClient.DefaultAPI.ApiAccountsSummaryGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts summary: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("no response from API")
	}
	defer r.Body.Close()

	var accountBalances []models.ExternalAccount
	for _, account := range resp.Data.GetProducts() {
		externalAccountName := AccountName(&account)
		primaryBalances := account.GetPrimaryBalances()
		if len(primaryBalances) == 0 {
			return nil, fmt.Errorf("no primary balances found for account %s", externalAccountName)
		}
		balance := primaryBalances[0]
		amount := models.Amount{
			Value:    fmt.Sprintf("%.2f", balance.GetAmount()),
			Currency: balance.GetCurrencyCode(),
		}
		accountBalances = append(accountBalances, models.ExternalAccount{
			Name:    externalAccountName,
			Balance: amount,
		})
	}
	return accountBalances, nil
}
