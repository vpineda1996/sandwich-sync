package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vpineda1996/sandwich-sync/pkg/models"

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reference_number TEXT,
		activity_type TEXT,
		amount_value TEXT,
		amount_currency TEXT,
		activity_status TEXT,
		activity_category TEXT,
		activity_classification TEXT,
		card_number TEXT,
		merchant_name TEXT,
		merchant_category_code TEXT,
		merchant_category_description TEXT,
		merchant_category TEXT,
		merchant_city TEXT,
		merchant_state_province TEXT,
		merchant_postal_code TEXT,
		merchant_country_code TEXT,
		transaction_date TEXT,
		activity_category_code TEXT,
		customer_id TEXT,
		posted_date TEXT,
		name_on_card TEXT,
		lunchmoney_id INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	return nil
}

// UpdateTransaction updates an existing transaction in the database
func (db *DB) UpdateTransaction(tx *models.Transaction) error {
	query := `
	UPDATE transactions
	SET 
		activity_type = ?, amount_value = ?, amount_currency = ?, activity_status = ?, 
		activity_category = ?, activity_classification = ?, card_number = ?, merchant_name = ?, 
		merchant_category_code = ?, merchant_category_description = ?, merchant_category = ?, 
		merchant_city = ?, merchant_state_province = ?, merchant_postal_code = ?, 
		merchant_country_code = ?, transaction_date = ?, activity_category_code = ?, 
		customer_id = ?, posted_date = ?, name_on_card = ?
		` + func() string {
		if tx.LunchMoneyID != 0 {
			return `, lunchmoney_id = ? `
		}
		return ``
	}() + `
	WHERE reference_number = ?
	`

	args := []interface{}{
		tx.ActivityType,
		tx.Amount.Value,
		tx.Amount.Currency,
		tx.ActivityStatus,
		tx.ActivityCategory,
		tx.ActivityClassification,
		tx.CardNumber,
		tx.Merchant.Name,
		tx.Merchant.CategoryCode,
		tx.Merchant.CategoryDescription,
		tx.Merchant.Category,
		tx.Merchant.Address.City,
		tx.Merchant.Address.StateProvince,
		tx.Merchant.Address.PostalCode,
		tx.Merchant.Address.CountryCode,
		tx.Date,
		tx.ActivityCategoryCode,
		tx.CustomerID,
		tx.PostedDate,
		tx.Name.NameOnCard,
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
func (db *DB) SaveTransaction(tx *models.Transaction) error {
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
		reference_number, activity_type, amount_value, amount_currency,
		activity_status, activity_category, activity_classification, card_number,
		merchant_name, merchant_category_code, merchant_category_description, merchant_category,
		merchant_city, merchant_state_province, merchant_postal_code, merchant_country_code,
		transaction_date, activity_category_code, customer_id, posted_date, name_on_card, lunchmoney_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(
		query,
		tx.ReferenceNumber,
		tx.ActivityType,
		tx.Amount.Value,
		tx.Amount.Currency,
		tx.ActivityStatus,
		tx.ActivityCategory,
		tx.ActivityClassification,
		tx.CardNumber,
		tx.Merchant.Name,
		tx.Merchant.CategoryCode,
		tx.Merchant.CategoryDescription,
		tx.Merchant.Category,
		tx.Merchant.Address.City,
		tx.Merchant.Address.StateProvince,
		tx.Merchant.Address.PostalCode,
		tx.Merchant.Address.CountryCode,
		tx.Date,
		tx.ActivityCategoryCode,
		tx.CustomerID,
		tx.PostedDate,
		tx.Name.NameOnCard,
		tx.LunchMoneyID,
	)

	if err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	return nil
}

// GetTransactions retrieves all transactions from the database
func (db *DB) GetTransactions() ([]*models.Transaction, error) {
	query := `
	SELECT 
		reference_number, activity_type, amount_value, amount_currency,
		activity_status, activity_category, activity_classification, card_number,
		merchant_name, merchant_category_code, merchant_category_description, merchant_category,
		merchant_city, merchant_state_province, merchant_postal_code, merchant_country_code,
		transaction_date, activity_category_code, customer_id, posted_date, name_on_card, lunchmoney_id
	FROM transactions
	ORDER BY transaction_date DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		tx := &models.Transaction{
			Amount:   &models.Amount{},
			Merchant: &models.Merchant{Address: &models.Address{}},
			Name:     &models.Name{},
		}

		err := rows.Scan(
			&tx.ReferenceNumber,
			&tx.ActivityType,
			&tx.Amount.Value,
			&tx.Amount.Currency,
			&tx.ActivityStatus,
			&tx.ActivityCategory,
			&tx.ActivityClassification,
			&tx.CardNumber,
			&tx.Merchant.Name,
			&tx.Merchant.CategoryCode,
			&tx.Merchant.CategoryDescription,
			&tx.Merchant.Category,
			&tx.Merchant.Address.City,
			&tx.Merchant.Address.StateProvince,
			&tx.Merchant.Address.PostalCode,
			&tx.Merchant.Address.CountryCode,
			&tx.Date,
			&tx.ActivityCategoryCode,
			&tx.CustomerID,
			&tx.PostedDate,
			&tx.Name.NameOnCard,
			&tx.LunchMoneyID,
		)

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
func (db *DB) GetTransactionByReference(referenceNumber string) (*models.Transaction, error) {
	query := `
	SELECT
		reference_number, activity_type, amount_value, amount_currency,
		activity_status, activity_category, activity_classification, card_number,
		merchant_name, merchant_category_code, merchant_category_description, merchant_category,
		merchant_city, merchant_state_province, merchant_postal_code, merchant_country_code,
		transaction_date, activity_category_code, customer_id, posted_date, name_on_card, lunchmoney_id
	FROM transactions
	WHERE reference_number = ?
	LIMIT 1
	`

	tx := &models.Transaction{
		Amount:   &models.Amount{},
		Merchant: &models.Merchant{Address: &models.Address{}},
		Name:     &models.Name{},
	}

	err := db.QueryRow(query, referenceNumber).Scan(
		&tx.ReferenceNumber,
		&tx.ActivityType,
		&tx.Amount.Value,
		&tx.Amount.Currency,
		&tx.ActivityStatus,
		&tx.ActivityCategory,
		&tx.ActivityClassification,
		&tx.CardNumber,
		&tx.Merchant.Name,
		&tx.Merchant.CategoryCode,
		&tx.Merchant.CategoryDescription,
		&tx.Merchant.Category,
		&tx.Merchant.Address.City,
		&tx.Merchant.Address.StateProvince,
		&tx.Merchant.Address.PostalCode,
		&tx.Merchant.Address.CountryCode,
		&tx.Date,
		&tx.ActivityCategoryCode,
		&tx.CustomerID,
		&tx.PostedDate,
		&tx.Name.NameOnCard,
		&tx.LunchMoneyID,
	)

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
func (db *DB) AddManualTransaction(tx *models.Transaction) error {
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
