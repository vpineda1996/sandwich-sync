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
		is_plaid BOOLEAN
	)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create account_info table: %w", err)
	}
	return err
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
		balance_updated_at = CURRENT_TIMESTAMP
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
		lunchmoney_account_id, balance_value, balance_currency, balance_updated_at, is_plaid
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
