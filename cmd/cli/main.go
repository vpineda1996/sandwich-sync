package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vpineda1996/sandwich-sync/db"
	"github.com/vpineda1996/sandwich-sync/pkg/config"
	"github.com/vpineda1996/sandwich-sync/pkg/http/rogers"
	"github.com/vpineda1996/sandwich-sync/pkg/http/ws"
	"github.com/vpineda1996/sandwich-sync/pkg/models"
	"github.com/vpineda1996/sandwich-sync/pkg/parser"
	"github.com/vpineda1996/sandwich-sync/pkg/services"
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
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	defaultDBPath := filepath.Join(homeDir, ".lunchmoney", "transactions.db")

	// Initialize configuration
	if err := config.InitGlobalConfig("config.yaml"); err != nil {
		// Only print a warning if the file doesn't exist, as GetConfig will create it later
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load configuration: %v\n", err)
			fmt.Fprintf(os.Stderr, "A default configuration will be used.\n")
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
		Long:  `Start an interactive REPL for executing curl-like commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			runREPL()
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
}

func runREPL() {
	fmt.Println("Welcome to the Lunchmoney REPL!")
	fmt.Println("Type 'exit' or 'quit' to exit.")
	fmt.Println("Enter a curl-like command to fetch transactions.")
	fmt.Println()

	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP client
	client := rogers.NewCurlClient()

	// Start REPL
	scanner := bufio.NewScanner(os.Stdin)
	var multilineInput strings.Builder
	isMultiline := false

	for {
		if isMultiline {
			fmt.Print("... ")
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" && !isMultiline {
			continue
		}

		if (trimmedLine == "exit" || trimmedLine == "quit") && !isMultiline {
			break
		}

		if trimmedLine == "help" && !isMultiline {
			printHelp()
			continue
		}

		if trimmedLine == "config" && !isMultiline {
			showConfig()
			continue
		}

		if strings.HasPrefix(trimmedLine, "list") && !isMultiline {
			listTransactions(database)
			continue
		}

		if strings.HasPrefix(trimmedLine, "add") && !isMultiline {
			addTransaction(trimmedLine, database)
			continue
		}

		if strings.HasPrefix(trimmedLine, "sync") && !isMultiline {
			syncTransactions(database)
			continue
		}

		if strings.HasPrefix(trimmedLine, "fetch") && !isMultiline {
			fetchTransactions(database)
			continue
		}

		if (strings.HasPrefix(trimmedLine, "remove") || strings.HasPrefix(trimmedLine, "delete")) && !isMultiline {
			removeTransaction(trimmedLine, database)
			continue
		}

		// Handle multiline input
		if strings.HasSuffix(trimmedLine, "\\") {
			// Remove the trailing backslash and add to the buffer
			multilineInput.WriteString(trimmedLine[:len(trimmedLine)-1])
			multilineInput.WriteString(" ")
			isMultiline = true
			continue
		} else if isMultiline {
			// Add the last line and process the complete command
			multilineInput.WriteString(line)
			input := multilineInput.String()

			// Process curl command
			processCurlCommand(input, client, database)

			// Reset for next command
			multilineInput.Reset()
			isMultiline = false
		} else {
			// Single line command
			processCurlCommand(line, client, database)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

// TODO incorporate to DB, need to do some changes in client
func fetchTransactionsWs(database *db.DB) {
	client, err := ws.NewWealthsimpleClient(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Wealthsimple client: %v\n", err)
		return
	}

	transactions, err := client.FetchTransactions(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching transactions: %v\n", err)
		return
	}

	for _, tx := range transactions {
		fmt.Printf("%35s %15s %12s %s\n", tx.ReferenceNumber, tx.Date, tx.Amount.Value, tx.Merchant.Name)
	}
}

func fetchTransactions(database *db.DB) {
	deviceId, err := config.GetRogersDeviceId()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting Rogers device ID: %v\n", err)
		return
	}
	client := rogers.NewRogersBankClient(deviceId)

	username, password, err := config.GetRogersCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting Rogers credentials: %v\n", err)
		return
	}

	if err := client.Authenticate(context.Background(), username, password); err != nil {
		fmt.Fprintf(os.Stderr, "Error authenticating: %v\n", err)
		return
	}

	// Fetch transactions
	transactions, err := client.FetchTransactions(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching transactions: %v\n", err)
		return
	}

	for _, tx := range transactions {
		if tx, err := database.GetTransactionByReference(tx.ReferenceNumber); tx != nil && err == nil {
			fmt.Printf("Transaction %s already exists in the database\n", tx.ReferenceNumber)
			continue
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking transaction: %v\n", err)
			continue
		}

		if err := database.SaveTransaction(&tx); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving transaction: %v\n", err)
			continue
		}
		fmt.Printf("Transaction %s saved successfully\n", tx.ReferenceNumber)
	}
}

func syncTransactions(database *db.DB) {
	// Get the API key from the configuration
	apiKey, err := config.GetLunchMoneyAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting API key from config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please set your API key in config.yaml\n")
		return
	}

	lsyncer, err := services.NewLunchMoneySyncer(context.Background(), apiKey, database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating LunchMoney syncer: %v\n", err)
		return
	}

	err = lsyncer.SyncTransactions(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error syncing transactions: %v\n", err)
		return
	}
}

func processCurlCommand(input string, client *rogers.CurlClient, database *db.DB) {
	// Parse curl command
	cmd, err := parser.ParseCurlCommand(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing curl command: %v\n", err)
		return
	}

	fmt.Println("Parsed command:")
	fmt.Println(cmd)

	// Add cookies to headers
	if len(cmd.Cookies) > 0 {
		var cookieStr strings.Builder
		first := true
		for key, value := range cmd.Cookies {
			if !first {
				cookieStr.WriteString("; ")
			}
			cookieStr.WriteString(key)
			cookieStr.WriteString("=")
			cookieStr.WriteString(value)
			first = false
		}
		cmd.Headers["Cookie"] = cookieStr.String()
	}

	// Fetch transactions
	fmt.Println("Fetching transactions...")
	transactions, err := client.FetchTransactions(cmd.URL, cmd.Headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching transactions: %v\n", err)
		return
	}

	fmt.Printf("Fetched %d transactions\n", len(transactions))

	// Save transactions to database
	for _, tx := range transactions {
		if err := database.SaveTransaction(&tx); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving transaction: %v\n", err)
			continue
		}
	}

	fmt.Printf("Saved %d transactions to database\n", len(transactions))
}

func listTransactions(database *db.DB) {
	transactions, err := database.GetTransactions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching transactions: %v\n", err)
		return
	}

	if len(transactions) == 0 {
		fmt.Println("No transactions found")
		return
	}

	fmt.Printf("Found %d transactions:\n\n", len(transactions))
	fmt.Printf("%-30s %-15s %-30s %-15s %-15s\n", "Reference Number", "Amount", "Merchant Name", "Date", "LunchMoney ID")
	fmt.Println(strings.Repeat("-", 110))
	for _, tx := range transactions {
		fmt.Printf("%-30s %-15s %-30s %-15s %-15d\n",
			tx.ReferenceNumber,
			tx.Amount.Value+" "+tx.Amount.Currency,
			tx.Merchant.Name,
			tx.Date,
			tx.LunchMoneyID)
	}
}

func addTransaction(input string, database *db.DB) {
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
	tx := &models.Transaction{
		ReferenceNumber:        referenceNumber,
		ActivityType:           "TRANS",
		Amount:                 &models.Amount{Value: amountValue, Currency: currency},
		ActivityStatus:         "APPROVED",
		ActivityCategory:       category,
		ActivityClassification: "PURCHASE",
		CardNumber:             "************0000", // Masked card number
		Merchant: &models.Merchant{
			Name:     merchantName,
			Category: category,
			Address:  &models.Address{},
		},
		Date:                 date,
		ActivityCategoryCode: "0001",
		CustomerID:           "MANUAL",
		PostedDate:           date,
		Name:                 &models.Name{NameOnCard: "MANUAL ENTRY"},
	}

	// Save transaction
	if err := database.AddManualTransaction(tx); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding transaction: %v\n", err)
		return
	}

	fmt.Printf("Transaction %s added successfully\n", referenceNumber)
}

func removeTransaction(input string, database *db.DB) {
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
	if err := database.RemoveTransaction(referenceNumber); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing transaction: %v\n", err)
		return
	}

	fmt.Printf("Transaction %s removed successfully\n", referenceNumber)
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                 - Show this help message")
	fmt.Println("  config               - Show the current configuration")
	fmt.Println("  list                 - List all transactions in the database")
	fmt.Println("  fetch                - Fetch transactions from Rogers Bank")
	fmt.Println("  sync                 - Sync database with LunchMoney API")
	fmt.Println("  add <ref> <amount> <currency> <merchant> <date> [<category>]")
	fmt.Println("                       - Add a transaction manually")
	fmt.Println("  remove <ref>         - Remove a transaction by reference number")
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
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
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
