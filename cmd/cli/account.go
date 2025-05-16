package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (r *replState) handleLunchMoneyAccounts(input string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		fmt.Println("Invalid account command format.")
		fmt.Println("Usage: account <list|disable>")
		return
	}

	if parts[1] == "list" || parts[1] == "l" {
		accounts, err := r.db.GetAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Error fetching accounts")
			return
		}
		lmAccounts, err := r.lmSyncer.GetClient().ListAccounts(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Error fetching accounts")
			return
		}
		lmAccountsMap := lo.SliceToMap(lmAccounts, func(account models.LunchMoneyAccount) (int64, models.LunchMoneyAccount) {
			return account.LunchMoneyId, account
		})
		for i := range accounts {
			if account, ok := lmAccountsMap[accounts[i].LunchMoneyId]; ok {
				accounts[i].Name = account.Name
				accounts[i].DisplayName = account.DisplayName
			}
		}

		if len(accounts) == 0 {
			fmt.Println("No accounts found")
			return
		}

		fmt.Printf("Found %d accounts:\n\n", len(accounts))
		fmt.Printf("%-10s %-30s %15s %-15s %-10s %-10s\n", "LM ID", "Account Name", "Balance", "Currency", "Sync Trns", "Sync Balance")
		fmt.Println(strings.Repeat("-", 115))
		for _, account := range accounts {
			fmt.Printf("%-10d %-30s %15s %-15s %-10t %-10t\n",
				account.LunchMoneyId,
				account.DisplayName[:min(30, len(account.DisplayName))],
				account.Balance.Value[:min(15, len(account.Balance.Value))],
				account.Balance.Currency,
				account.SyncStrategy&models.SyncOptionTransactions != 0,
				account.SyncStrategy&models.SyncOptionBalance != 0)
		}
	} else if parts[1] == "disable" || parts[1] == "d" {
		if len(parts) < 4 {
			fmt.Println("Usage: account disable <transaction|balance> <lunchmoney_id>")
			return
		}

		var syncOption models.SyncOption
		switch strings.ToLower(parts[2]) {
		case "transaction", "t":
			syncOption = models.SyncOptionTransactions
		case "balance", "b":
			syncOption = models.SyncOptionBalance
		default:
			fmt.Println("Invalid sync option. Supported options are: transaction, balance")
			return
		}

		lunchMoneyId := parts[3]
		if err := r.db.DisableSyncOptions(lunchMoneyId, syncOption); err != nil {
			log.Error().Err(err).Msg("Error disabling account")
			return
		}
		log.Info().Str("account", lunchMoneyId).Msg("Account disabled successfully")
	} else {
		fmt.Println("Unknown command. Supported commands are: list, disable")
	}
}
