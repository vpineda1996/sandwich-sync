package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vpnda/sandwich-sync/pkg/models"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Initialize creates the necessary tables if they don't exist
func (db *DB) Initialize() error {
	query := `
	CREATE TABLE IF NOT EXISTS transactions (
		reference_number TEXT PRIMARY KEY,
		amount_value TEXT,
		amount_currency TEXT,
		merchant_name TEXT,
		merchant_category_code TEXT,
		merchant_city TEXT,
		merchant_state_province TEXT,
		transaction_date TEXT,
		activity_category_code TEXT,
		customer_id TEXT,
		posted_date TEXT,
		name_on_card TEXT,
		source_account_name TEXT,
		lunchmoney_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	// Add source account name column if it doesn't exist
	query = `
	select count(*) from
	pragma_table_info('transactions')
	where name='source_account_name';
	`
	var count int
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for source_account_name column: %w", err)
	}

	if count == 0 {
		query = `
		ALTER TABLE transactions ADD COLUMN source_account_name TEXT;
		`
		_, err = db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to add source_account_name column: %w", err)
		}
	}

	query = `
	CREATE TABLE IF NOT EXISTS account_mappings (
		external_name TEXT PRIMARY KEY,
		lunchmoney_account_id TEXT,
		is_plaid BOOLEAN
	)
	`
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create account_mappings table: %w", err)
	}

	return nil
}

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

// UpdateTransaction updates an existing transaction in the database
func (db *DB) UpdateTransaction(tx *models.TransactionWithAccount) error {
	query := `
	UPDATE transactions
	SET 
		amount_value = ?, amount_currency = ?, merchant_name = ?, 
		merchant_category_code = ?,
		merchant_city = ?, merchant_state_province = ?, 
		transaction_date = ?, posted_date = ?, source_account_name = ?
		` + func() string {
		if tx.LunchMoneyID != 0 {
			return `, lunchmoney_id = ? `
		}
		return ``
	}() + `
	WHERE reference_number = ?
	`

	args := []interface{}{
		tx.Amount.Value,
		tx.Amount.Currency,
		tx.Merchant.Name,
		tx.Merchant.CategoryCode,
		tx.Merchant.Address.City,
		tx.Merchant.Address.StateProvince,
		tx.Date,
		tx.PostedDate,
		tx.SourceAccountName,
	}

	if tx.LunchMoneyID != 0 {
		args = append(args, tx.LunchMoneyID)
	}

	args = append(args, tx.ReferenceNumber)

	result, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no transaction found with reference number: %s", tx.ReferenceNumber)
	}

	return nil
}

// SaveTransaction saves a transaction to the database
func (db *DB) SaveTransaction(tx *models.TransactionWithAccount) error {
	// Check if a transaction with the same reference number already exists
	existingTx, err := db.GetTransactionByReference(tx.ReferenceNumber)
	if err != nil {
		return fmt.Errorf("failed to check for existing transaction: %w", err)
	}

	if existingTx != nil {
		// Update the existing transaction
		return db.UpdateTransaction(tx)
	}

	if tx.Merchant == nil {
		tx.Merchant = &models.Merchant{}
	}

	if tx.Merchant.Address == nil {
		tx.Merchant.Address = &models.Address{}
	}

	// Insert a new transaction
	query := `
	INSERT INTO transactions (
		reference_number, amount_value, amount_currency,
		merchant_name, merchant_category_code,
		merchant_city, merchant_state_province,
		transaction_date, posted_date, source_account_name, lunchmoney_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(
		query,
		tx.ReferenceNumber,
		tx.Amount.Value,
		tx.Amount.Currency,
		tx.Merchant.Name,
		tx.Merchant.CategoryCode,
		tx.Merchant.Address.City,
		tx.Merchant.Address.StateProvince,
		tx.Date,
		tx.PostedDate,
		tx.SourceAccountName,
		tx.LunchMoneyID,
	)

	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}

// GetTransactions retrieves all transactions from the database
func (db *DB) GetTransactions() ([]*models.TransactionWithAccount, error) {
	query := `
	SELECT 
		reference_number, amount_value, amount_currency,
		merchant_name, merchant_category_code,
		merchant_city, merchant_state_province,
		transaction_date, posted_date, source_account_name, lunchmoney_id
	FROM transactions
	ORDER BY transaction_date DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.TransactionWithAccount
	for rows.Next() {
		tx := &models.TransactionWithAccount{
			Transaction: models.Transaction{
				Amount:   models.Amount{},
				Merchant: &models.Merchant{Address: &models.Address{}},
			},
		}

		var nullStr sql.NullString

		err := rows.Scan(
			&tx.ReferenceNumber,
			&tx.Amount.Value,
			&tx.Amount.Currency,
			&tx.Merchant.Name,
			&tx.Merchant.CategoryCode,
			&tx.Merchant.Address.City,
			&tx.Merchant.Address.StateProvince,
			&tx.Date,
			&tx.PostedDate,
			&nullStr,
			&tx.LunchMoneyID,
		)

		if nullStr.Valid {
			tx.SourceAccountName = nullStr.String
		}

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetTransactionByReference retrieves a transaction by its reference number
func (db *DB) GetTransactionByReference(referenceNumber string) (*models.TransactionWithAccount, error) {
	query := `
	SELECT
		reference_number, amount_value, amount_currency,
		merchant_name, merchant_category_code,
		merchant_city, merchant_state_province,
		transaction_date, posted_date, source_account_name, lunchmoney_id
	FROM transactions
	WHERE reference_number = ?
	LIMIT 1
	`

	tx := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			Amount:   models.Amount{},
			Merchant: &models.Merchant{Address: &models.Address{}},
		},
	}

	var nullStr sql.NullString

	err := db.QueryRow(query, referenceNumber).Scan(
		&tx.ReferenceNumber,
		&tx.Amount.Value,
		&tx.Amount.Currency,
		&tx.Merchant.Name,
		&tx.Merchant.CategoryCode,
		&tx.Merchant.Address.City,
		&tx.Merchant.Address.StateProvince,
		&tx.Date,
		&tx.PostedDate,
		&nullStr,
		&tx.LunchMoneyID,
	)

	if nullStr.Valid {
		tx.SourceAccountName = nullStr.String
	}

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return tx, nil
}

// RemoveTransaction removes a transaction by its reference number
func (db *DB) RemoveTransaction(referenceNumber string) error {
	query := `DELETE FROM transactions WHERE reference_number = ?`

	result, err := db.Exec(query, referenceNumber)
	if err != nil {
		return fmt.Errorf("failed to remove transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no transaction found with reference number: %s", referenceNumber)
	}

	return nil
}

// AddManualTransaction adds a manually created transaction to the database
func (db *DB) AddManualTransaction(tx *models.TransactionWithAccount) error {
	// Check if transaction with this reference number already exists
	existingTx, err := db.GetTransactionByReference(tx.ReferenceNumber)
	if err != nil {
		return fmt.Errorf("failed to check for existing transaction: %w", err)
	}

	if existingTx != nil {
		return fmt.Errorf("transaction with reference number %s already exists", tx.ReferenceNumber)
	}

	// Save the transaction
	return db.SaveTransaction(tx)
}
