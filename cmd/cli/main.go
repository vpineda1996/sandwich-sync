package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/vpnda/sandwich-sync/db"
	"github.com/vpnda/sandwich-sync/pkg/config"
	"github.com/vpnda/sandwich-sync/pkg/models"
	"github.com/vpnda/sandwich-sync/pkg/services"
)

var (
	dbPath  string
	rootCmd *cobra.Command
)

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Msg("Error getting home directory")
		os.Exit(1)
	}

	defaultDBPath := filepath.Join(homeDir, ".lunchmoney", "transactions.db")

	// Initialize configuration
	if err := config.InitGlobalConfig("config.yaml"); err != nil {
		// Only print a warning if the file doesn't exist, as GetConfig will create it later
		if !os.IsNotExist(err) {
			log.Warn().Err(err).Msg("Failed to load configuration")
			log.Warn().Msg("A default configuration will be used")
		}
	}

	rootCmd = &cobra.Command{
		Use:   "lunchmoney",
		Short: "A CLI tool for fetching and storing transactions",
		Long:  `A CLI tool that fetches transactions from an API and stores them in a SQLite database.`,
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "Path to the SQLite database")

	replCmd := &cobra.Command{
		Use:   "repl",
		Short: "Start an interactive REPL",
		Long:  `Start an interactive REPL for executing commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			runREPL(initReplState(cmd.Context()))
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Show the current configuration",
		Long:  `Show the current configuration loaded from config.yaml.`,
		Run: func(cmd *cobra.Command, args []string) {
			showConfig()
		},
	}

	rootCmd.AddCommand(replCmd, configCmd)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func initReplState(ctx context.Context) replState {
	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to database")
		os.Exit(1)
	}

	// Get the API key from the configuration
	apiKey, err := config.GetLunchMoneyAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("Error getting API key from config")
		log.Error().Msg("Please set your API key in config.yaml")
		os.Exit(1)
	}

	lsyncer, err := services.NewLunchMoneySyncer(ctx, apiKey, database)
	if err != nil {
		log.Error().Err(err).Msg("Error creating LunchMoney syncer")
		os.Exit(1)
	}
	return replState{
		db:       database,
		lmSyncer: lsyncer,
	}
}

type replState struct {
	db       db.DBInterface
	lmSyncer *services.LunchMoneySyncer
}

func runREPL(state replState) {
	fmt.Println("Welcome to the Lunchmoney REPL!")
	fmt.Println("Type 'exit' or 'quit' to exit.")
	fmt.Println("Enter a command to pull/push transactions.")
	fmt.Println()

	// Close the database once you are done
	defer state.db.Close()

	if err := state.db.Initialize(); err != nil {
		log.Error().Err(err).Msg("Error initializing database")
		os.Exit(1)
	}

	// Start REPL
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue
		}

		if trimmedLine == "exit" || trimmedLine == "quit" {
			break
		}

		if trimmedLine == "help" {
			printHelp()
			continue
		}

		if trimmedLine == "config" {
			showConfig()
			continue
		}

		if strings.HasPrefix(trimmedLine, "list") {
			state.listTransactions()
			continue
		}

		if strings.HasPrefix(trimmedLine, "account") {
			state.handleLunchMoneyAccounts(trimmedLine)
			continue
		}

		if strings.HasPrefix(trimmedLine, "add") {
			state.addTransaction(trimmedLine)
			continue
		}

		if strings.HasPrefix(trimmedLine, "sync") {
			state.syncState()
			continue
		}

		if strings.HasPrefix(trimmedLine, "fetch") {
			state.processTransactionFetch(trimmedLine)
			continue
		}

		if strings.HasPrefix(trimmedLine, "remove") || strings.HasPrefix(trimmedLine, "delete") {
			state.removeTransaction(trimmedLine)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Msg("Error reading input")
	}
}

func (r *replState) syncState() {
	err := r.lmSyncer.SyncTransactions(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error syncing transactions")
		return
	}

	err = r.lmSyncer.SyncBalances(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error syncing balances")
		return
	}
}

func (r *replState) listTransactions() {
	transactions, err := r.db.GetTransactions()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching transactions")
		return
	}

	if len(transactions) == 0 {
		fmt.Println("No transactions found")
		return
	}

	fmt.Printf("Found %d transactions:\n\n", len(transactions))
	fmt.Printf("%-20s %-30s %-15s %-30s %-15s %-15s\n", "SourceAccount", "Reference Number", "Amount", "Merchant Name", "Date", "LunchMoney ID")
	fmt.Println(strings.Repeat("-", 130))
	for _, tx := range transactions {
		fmt.Printf("%-20s %-30s %-15s %-30s %-15s %-15d\n",
			tx.SourceAccountName[:min(20, len(tx.SourceAccountName))],
			tx.ReferenceNumber[:min(30, len(tx.ReferenceNumber))],
			tx.Amount.Value+" "+tx.Amount.Currency,
			tx.Merchant.Name[:min(30, len(tx.Merchant.Name))],
			tx.Date,
			tx.LunchMoneyID)
	}
}

func (r *replState) addTransaction(input string) {
	// Parse the add command
	// Format: add <reference_number> <amount> <currency> <merchant_name> <date> [<category>]
	parts := strings.Fields(input)
	if len(parts) < 6 {
		fmt.Println("Invalid add command format.")
		fmt.Println("Usage: add <reference_number> <amount> <currency> <merchant_name> <date> [<category>]")
		fmt.Println("Example: add TX123456 25.99 USD \"Coffee Shop\" 2025-04-29 FOOD")
		return
	}

	// Extract parameters
	referenceNumber := parts[1]
	amountValue := parts[2]
	currency := parts[3]

	// Merchant name might contain spaces and be quoted
	merchantName := parts[4]
	if strings.HasPrefix(merchantName, "\"") && !strings.HasSuffix(merchantName, "\"") {
		// Find the closing quote
		merchantNameParts := []string{merchantName}
		for i := 5; i < len(parts); i++ {
			merchantNameParts = append(merchantNameParts, parts[i])
			if strings.HasSuffix(parts[i], "\"") {
				parts = append(parts[:4], parts[i+1:]...)
				break
			}
		}
		merchantName = strings.Join(merchantNameParts, " ")
		merchantName = strings.Trim(merchantName, "\"")
	} else {
		merchantName = strings.Trim(merchantName, "\"")
	}

	// Continue with remaining parameters
	if len(parts) < 6 {
		fmt.Println("Invalid add command format after parsing merchant name.")
		fmt.Println("Usage: add <reference_number> <amount> <currency> <merchant_name> <date> [<category>]")
		return
	}

	date := parts[5]

	// Optional category
	category := "PURCHASE"
	if len(parts) >= 7 {
		category = parts[6]
	}

	// Create transaction
	tx := &models.TransactionWithAccount{
		Transaction: models.Transaction{
			ReferenceNumber: referenceNumber,
			Amount:          models.Amount{Value: amountValue, Currency: currency},
			Merchant: &models.Merchant{
				Name:         merchantName,
				CategoryCode: category,
				Address:      &models.Address{},
			},
			Date:       date,
			PostedDate: date,
		},
		SourceAccountName: "Manual Entry",
	}

	// Save transaction
	if err := r.db.AddManualTransaction(tx); err != nil {
		log.Error().Err(err).Msg("Error adding transaction")
		return
	}

	log.Info().Str("transaction", referenceNumber).Msg("Transaction added successfully")
}

func (r *replState) removeTransaction(input string) {
	// Parse the remove command
	// Format: remove <reference_number>
	parts := strings.Fields(input)
	if len(parts) != 2 {
		fmt.Println("Invalid remove command format.")
		fmt.Println("Usage: remove <reference_number>")
		fmt.Println("Example: remove TX123456")
		return
	}

	// Extract reference number
	referenceNumber := parts[1]

	// Remove transaction
	if err := r.db.RemoveTransaction(referenceNumber); err != nil {
		log.Error().Err(err).Msg("Error removing transaction")
		return
	}

	log.Info().Str("transaction", referenceNumber).Msg("Transaction removed successfully")
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                 - Show this help message")
	fmt.Println("  config               - Show the current configuration")
	fmt.Println("  list                 - List all transactions in the database")
	fmt.Println("  fetch <type>         - Fetch transactions from either 'wealthsimple', ")
	fmt.Println("                         'rogers', or 'scotia'")
	fmt.Println("  sync                 - Sync database with LunchMoney API")
	fmt.Println("  add <ref> <amount> <currency> <merchant> <date> [<category>]")
	fmt.Println("                       - Add a transaction manually")
	fmt.Println("  remove <ref>         - Remove a transaction by reference number")
	fmt.Println("  account list         - List all accounts with balances and sync status")
	fmt.Println("  account disable <id> - Disable syncing for an account by its LunchMoney ID")
	fmt.Println("  exit, quit           - Exit the REPL")
	fmt.Println("  curl [command]       - Execute a curl-like command to fetch transactions")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  The application uses a config.yaml file in the current directory.")
	fmt.Println("  Make sure to set your lunchMoneyApiKey in this file before using the sync command.")
}

// showConfig displays the current configuration
func showConfig() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("Error loading configuration")
		return
	}

	fmt.Println("Current Configuration:")
	fmt.Println("----------------------")

	// Display the API key (masked for security)
	apiKey := cfg.LunchMoneyAPIKey
	maskedKey := ""
	if apiKey != "" {
		// Show only the first 4 and last 4 characters of the API key
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
		} else {
			maskedKey = strings.Repeat("*", len(apiKey))
		}
		fmt.Printf("Lunch Money API Key: %s\n", maskedKey)
	} else {
		fmt.Println("Lunch Money API Key: Not set")
		fmt.Println("\nPlease set your API key in config.yaml to use the sync command.")
		fmt.Println("You can get your API key from https://my.lunchmoney.app/developers")
	}
}
