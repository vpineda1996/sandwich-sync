package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/config"
	"github.com/vpnda/sandwich-sync/pkg/models"
	"github.com/vpnda/wsfetch/pkg/auth/types"
	"github.com/vpnda/wsfetch/pkg/base"
	"github.com/vpnda/wsfetch/pkg/client"
)

type WealthsimpleClient struct {
	c client.Client
}

func NewWealthsimpleClient(ctx context.Context) (*WealthsimpleClient, error) {
	prevSession, err := config.GetWealthsimplePrevSession()
	var authClient *base.Wealthsimple
	if err != nil {
		log.Info().Err(err).Msg("No previous session found, using password")
		username, password, err := config.GetWealthsimpleCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to get username/password: %w", err)
		}
		authClient = base.DefaultAuthClient(types.PasswordCredentials{
			Username: username,
			Password: password,
		})
	} else {
		log.Info().Msg("Using previous session")

		var sess types.Session
		err := json.Unmarshal([]byte(prevSession), &sess)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal previous session: %w", err)
		}
		authClient = base.AuthClientFromSession(&sess)
	}

	newSess, err := authClient.Fetcher.GetSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new session: %w", err)
	}

	b, err := json.Marshal(newSess)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal new session: %w", err)
	}

	err = config.SetWealthsimplePrevSession(string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to set new session: %w", err)
	}

	c, err := client.NewClient(ctx, authClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &WealthsimpleClient{
		c: c,
	}, nil
}

// FetchTransactions implements http.TransactionFetcher.
func (w *WealthsimpleClient) FetchTransactions(ctx context.Context) ([]models.TransactionWithAccount, error) {
	accounts, err := w.c.GetAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	log.Info().Msgf("Found %d accounts", len(accounts))
	var transactions []models.TransactionWithAccount
	from := time.Now().Add(-30 * 24 * time.Hour)
	until := time.Now()

	for _, account := range accounts {
		activity, err := w.c.GetActivities(ctx, []client.AccountId{client.AccountId(account.Id)}, &from, &until)
		if err != nil {
			return nil, fmt.Errorf("failed to get transactions: %w", err)
		}
		trns, ok := activity[client.AccountId(account.Id)]
		if !ok {
			log.Info().Msgf("No transactions found for account %s", account.Id)
			continue
		}
		log.Info().Msgf("Found %d transactions for account %s", len(trns), account.Id)

		for _, trn := range trns {
			desc, err := client.GetActivityDescription(ctx, w.c, &trn)
			if err != nil {
				return nil, fmt.Errorf("failed to get transaction description: %w", err)
			}

			transactions = append(transactions, models.TransactionWithAccount{
				Transaction: models.Transaction{
					ReferenceNumber: *trn.CanonicalId,
					Merchant: &models.Merchant{
						Name: desc,
					},
					Amount: models.Amount{
						Value:    client.GetFormattedAmount(&trn),
						Currency: *trn.Currency,
					},
					Date: trn.OccurredAt.Format(time.DateOnly),
				},
				SourceAccountName: account.Id,
			})
		}
	}

	log.Info().Msgf("Found %d transactions", len(transactions))
	return transactions, nil
}
