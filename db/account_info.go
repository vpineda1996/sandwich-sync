package db

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (db *DB) createAccountInfoTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS account_info (
		lunchmoney_account_id INTEGER PRIMARY KEY,
		balance_value TEXT,
		balance_currency TEXT,
		balance_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_plaid BOOLEAN DEFAULT false,
		should_sync BOOLEAN DEFAULT true
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create account_info table: %w", err)
	}

	return err
}

func (db *DB) DisableAccountSync(lunchMoneyId string) error {
	query := `
	UPDATE account_info
	SET should_sync = false
	WHERE lunchmoney_account_id = ?
	`
	_, err := db.Exec(query, lunchMoneyId)
	if err != nil {
		return fmt.Errorf("failed to disable account sync: %w", err)
	}
	return nil
}

func (db *DB) IsAccountSyncEnabled(lunchMoneyId int64) (bool, error) {
	query := `
	SELECT should_sync FROM account_info WHERE lunchmoney_account_id = ?
	`
	row := db.QueryRow(query, lunchMoneyId)
	var shouldSync bool
	err := row.Scan(&shouldSync)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to query account sync status: %w", err)
	}
	return shouldSync, nil
}

// UpsertAccountBalance saves the balance for a given external account ID
func (db *DB) UpsertAccountBalance(externalId string, balance models.Amount) error {
	// first check if we've mapped the account
	query := `
	SELECT CAST(lunchmoney_account_id as INTEGER) FROM account_mappings WHERE external_name = ?
	`
	row := db.QueryRow(query, externalId)
	var lunchmoneyAccountId int64

	err := row.Scan(&lunchmoneyAccountId)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Info().Str("external_id", externalId).Msg("account not mapped, skip balance update")
			return nil
		}
		return fmt.Errorf("failed to query account mapping: %w", err)
	}

	// if we have, upsert the balance
	query = `
	INSERT INTO account_info (lunchmoney_account_id, balance_value, balance_currency, balance_updated_at, is_plaid)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP, false)
	ON CONFLICT(lunchmoney_account_id) 
	DO UPDATE SET 
		balance_value = excluded.balance_value,
		balance_currency = excluded.balance_currency,
		balance_updated_at = CURRENT_TIMESTAMP,
		is_plaid = excluded.is_plaid
	`

	_, err = db.Exec(query, lunchmoneyAccountId, balance.Value, balance.Currency)
	if err != nil {
		return fmt.Errorf("failed to upsert account balance: %w", err)
	}

	return nil
}

func (db *DB) GetAccounts() ([]models.LunchMoneyAccount, error) {
	query := `
	SELECT 
		lunchmoney_account_id, balance_value, balance_currency, balance_updated_at, is_plaid, should_sync
	FROM account_info
	WHERE lunchmoney_account_id >= 0
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()
	var accounts []models.LunchMoneyAccount
	for rows.Next() {
		var account models.LunchMoneyAccount
		err := rows.Scan(
			&account.LunchMoneyId,
			&account.Balance.Value,
			&account.Balance.Currency,
			&account.BalanceLastUpdated,
			&account.IsPlaid,
			&account.ShouldSync,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over accounts: %w", err)
	}
	return accounts, nil
}
