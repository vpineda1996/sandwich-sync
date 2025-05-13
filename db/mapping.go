package db

import (
	"database/sql"
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (db *DB) createAccountMappingsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS account_mappings (
		external_name TEXT PRIMARY KEY,
		lunchmoney_account_id TEXT,
		is_plaid BOOLEAN
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create account_mappings table: %w", err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS ignored_external_accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		external_name TEXT NOT NULL
	)
	`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create ignored_external_accounts table: %w", err)
	}
	return nil
}

// UpsertAccountMapping saves the mapping for a given external account ID
// and LunchMoney account ID. If the mapping already exists, it updates the
// LunchMoney account ID and is_plaid flag. If the mapping does not exist,
// it creates a new mapping.
// If the LunchMoney account ID is -1, it means that the account is ignored.
// In this case, we insert the account into the ignored_external_accounts table
// and set the LunchMoney account ID to -1 * the ID of the ignored account.
func (db *DB) UpsertAccountMapping(am *models.AccountMapping) error {
	if am.LunchMoneyId == -1 {
		query := `
		INSERT INTO ignored_external_accounts (external_name)
		VALUES (?)
		`
		_, err := db.Exec(query, am.ExternalName)
		if err != nil {
			return fmt.Errorf("failed to insert ignored account: %w", err)
		}

		query = `
		SELECT id FROM ignored_external_accounts WHERE external_name = ?
		`
		err = db.QueryRow(query, am.ExternalName).Scan(&am.LunchMoneyId)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("failed to get ignored account ID: %w", err)
			}
			return err
		}
		am.LunchMoneyId = -1 * am.LunchMoneyId
	}
	query := `
	INSERT INTO account_mappings (external_name, lunchmoney_account_id, is_plaid)
	VALUES (?, ?, ?)
	ON CONFLICT(external_name) DO UPDATE SET
		lunchmoney_account_id = excluded.lunchmoney_account_id,
		is_plaid = excluded.is_plaid
	`

	_, err := db.Exec(query, am.ExternalName, am.LunchMoneyId, am.IsPlaid)
	if err != nil {
		return fmt.Errorf("failed to upsert account mapping: %w", err)
	}

	return nil
}

func (db *DB) GetAccountMapping(externalId string) (*models.AccountMapping, error) {
	query := `
	SELECT 
		external_name, lunchmoney_account_id, is_plaid
	FROM account_mappings
	WHERE external_name = ?
	LIMIT 1
	`

	var am models.AccountMapping
	err := db.QueryRow(query, externalId).Scan(
		&am.ExternalName,
		&am.LunchMoneyId,
		&am.IsPlaid,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get account mapping: %w", err)
	}

	return &am, nil
}
