package db

import (
	"database/sql"
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/models"
)

func (db *DB) UpsertAccountMapping(am *models.AccountMapping) error {
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
