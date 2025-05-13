package rogers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	iface "github.com/vpnda/sandwich-sync/pkg/http"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (c *RogersBankClient) UpdateAccountBalances(ctx context.Context, balanceStorage iface.BalanceStorer) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("client is not authenticated")
	}

	detailReq, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(detailPath, c.accountId, c.customerId), nil)

	detailReq.Header = getCommonHeaders()
	detailResp, err := c.client.Do(detailReq)
	if err != nil {
		return fmt.Errorf("activity request failed: %w", err)
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
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if detail.CurrentBalance.Value == "" {
		return fmt.Errorf("empty balance value")
	}

	err = balanceStorage.UpsertAccountBalance(externalAccountName, models.Amount{
		Value:    detail.CurrentBalance.Value,
		Currency: detail.CurrentBalance.Currency,
	})
	return err
}
