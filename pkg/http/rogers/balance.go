package rogers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (c *RogersBankClient) FetchAccountBalances(ctx context.Context) ([]models.ExternalAccount, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("client is not authenticated")
	}

	detailReq, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(detailPath, c.accountId, c.customerId), nil)

	detailReq.Header = getCommonHeaders()
	detailResp, err := c.client.Do(detailReq)
	if err != nil {
		return nil, fmt.Errorf("activity request failed: %w", err)
	}
	defer detailResp.Body.Close()

	type detailResponse struct {
		CurrentBalance struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"currentBalance"`
	}

	var detail detailResponse
	if err := json.NewDecoder(detailResp.Body).Decode(&detail); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if detail.CurrentBalance.Value == "" {
		return nil, fmt.Errorf("empty balance value")
	}

	return []models.ExternalAccount{
		{
			Name: externalAccountName,
			Balance: models.Amount{
				Value:    detail.CurrentBalance.Value,
				Currency: detail.CurrentBalance.Currency,
			},
		},
	}, nil
}
