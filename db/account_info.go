package db

import (
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (db *DB) createAccountInfoTable() error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS account_info (
		lunchmoney_account_id INTEGER PRIMARY KEY,
		balance_value TEXT,
		balance_currency TEXT,
		balance_updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_plaid BOOLEAN DEFAULT false,
		sync_strategy INTEGER DEFAULT %d
	)`, models.AllSyncOption)

	_, err := db.Exec(query, models.AllSyncOption)
	if err != nil {
		return fmt.Errorf("failed to create account_info table: %w", err)
	}

	// Add source account name column if it doesn't exist
	query = `
	select count(*) from
	pragma_table_info('account_info')
	where name='sync_strategy';
	`
	var count int
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for sync_strategy column: %w", err)
	}

	if count == 0 {
		query = fmt.Sprintf(`
		ALTER TABLE account_info ADD COLUMN sync_strategy INTEGER DEFAULT %d;
		`, models.AllSyncOption)

		_, err = db.Exec(query, models.AllSyncOption)
		if err != nil {
			return fmt.Errorf("failed to add sync_strategy column: %w", err)
		}
	}

	return err
}

func (db *DB) DisableSyncOptions(lunchMoneyId string, syncOption models.SyncOption) error {
	query := `
	UPDATE account_info
	SET sync_strategy = sync_strategy & (~ ?)
	WHERE lunchmoney_account_id = ?
	`
	_, err := db.Exec(query, syncOption, lunchMoneyId)
	if err != nil {
		return fmt.Errorf("failed to disable account sync: %w", err)
	}
	return nil
}

func (db *DB) IsSyncOptionEnabled(lunchMoneyId int64, syncOption models.SyncOption) (bool, error) {
	query := `
	SELECT sync_strategy FROM account_info WHERE lunchmoney_account_id = ?
	`
	row := db.QueryRow(query, lunchMoneyId)
	var storedSyncOptions int64
	err := row.Scan(&storedSyncOptions)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to query account sync status: %w", err)
	}
	return (models.SyncOption(storedSyncOptions) & syncOption) > 0, nil
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
		lunchmoney_account_id, balance_value, balance_currency, 
		balance_updated_at, is_plaid, sync_strategy
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
			&account.SyncStrategy,
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
