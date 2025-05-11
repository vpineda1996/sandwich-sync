package rogers

import (
	"context"

	"github.com/vpnda/sandwich-sync/pkg/http"
)

// UpdateAccountBalances implements http.BalanceFetcher.
func (c *RogersBankClient) UpdateAccountBalances(ctx context.Context, balanceStorage http.BalanceStorer) error {
	// NOT IMPLEMENTED
	// TODO: Implement Rogers Bank balance fetching logic
	return nil
}
