package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/config"
	"github.com/vpnda/sandwich-sync/pkg/http"
	"github.com/vpnda/sandwich-sync/pkg/http/rogers"
	"github.com/vpnda/sandwich-sync/pkg/http/scotia"
	"github.com/vpnda/sandwich-sync/pkg/http/ws"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (r *replState) processTransactionFetch(trimmedLine string) {
	// Parse the fetch command
	parts := strings.Fields(trimmedLine)
	if len(parts) < 2 {
		fmt.Println("Invalid fetch command format.")
		fmt.Println("Usage: fetch <type>")
		fmt.Println("Example: fetch wealthsimple")
		return
	}

	fetchTypes := []string{parts[1]}
	if fetchTypes[0] == "all" {
		fetchTypes = []string{"wealthsimple", "rogers", "scotia"}
	}

	for _, fetfetchTypes := range fetchTypes {
		switch fetfetchTypes {
		case "wealthsimple":
			r.fetchTransactionsWs()
		case "rogers":
			r.fetchTransactionsRogers()
		case "scotia":
			r.fetchTransactionsScotia()
		default:
			fmt.Println("Unknown fetch type. Supported types are: wealthsimple, rogers, scotia, all")
		}
	}
}

func (r *replState) fetchTransactionsScotia() {
	client, err := scotia.NewScotiaClient()
	if err != nil {
		log.Error().Err(err).Msg("Error creating Scotia client")
		return
	}
	if err := client.AuthenticateDynamic(context.Background()); err != nil {
		log.Error().Err(err).Msg("Error authenticating Scotia client")
		return
	}
	r.syncFromFetcher(client)
}

func (r *replState) fetchTransactionsWs() {
	client, err := ws.NewWealthsimpleClient(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error creating Wealthsimple client")
		return
	}
	r.syncFromFetcher(client)
}

func (r *replState) fetchTransactionsRogers() {
	deviceId, err := config.GetRogersDeviceId()
	if err != nil {
		log.Error().Err(err).Msg("Error getting Rogers device ID")
		return
	}
	client := rogers.NewRogersBankClient(deviceId)

	username, password, err := config.GetRogersCredentials()
	if err != nil {
		log.Error().Err(err).Msg("Error getting Rogers credentials")
		return
	}

	if err := client.Authenticate(context.Background(), username, password); err != nil {
		log.Error().Err(err).Msg("Error authenticating Rogers client")
		return
	}
	r.syncFromFetcher(client)
}

func (r *replState) syncFromFetcher(client http.Fetcher) {
	accountBalances, err := client.FetchAccountBalances(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error fetching account balances")
		return
	}
	err = r.updateAccountBalances(accountBalances)
	if err != nil {
		log.Error().Err(err).Msg("Error updating account balances")
		return
	}

	// Fetch transactions
	transactions, err := client.FetchTransactions(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error fetching transactions")
		return
	}
	r.insertTransactionsToDb(transactions)
}

func (r *replState) updateAccountBalances(accountBalances []models.ExternalAccount) error {
	for _, account := range accountBalances {
		_, err := r.lmSyncer.GetAccountMapper().FindPossibleAccountForExternal(context.Background(), &account)
		if err != nil {
			log.Error().Err(err).Msg("Error finding account mapping")
			continue
		}
		if err := r.db.UpsertAccountBalance(account.Name, account.Balance); err != nil {
			log.Error().Err(err).Msg("Error updating account balance")
			return err
		}
		log.Info().Str("account", account.Name).Msg("Account balance updated successfully")
	}
	return nil
}

func (r *replState) insertTransactionsToDb(transactions []models.TransactionWithAccount) {
	inserted, skipped := 0, 0
	for _, tx := range transactions {
		if tx, err := r.db.GetTransactionByReference(tx.ReferenceNumber); tx != nil && err == nil {
			skipped++
			continue
		} else if err != nil {
			log.Error().Err(err).Msg("Error checking transaction")
			skipped++
			continue
		}

		if err := r.db.SaveTransaction(&tx); err != nil {
			log.Error().Err(err).Msg("Error saving transaction")
			continue
		}
		log.Info().Str("transaction", tx.ReferenceNumber).Msg("Transaction saved successfully")
		inserted++
	}
	log.Info().Int("inserted", inserted).Int("skipped", skipped).Msg("Transactions processed")
}
