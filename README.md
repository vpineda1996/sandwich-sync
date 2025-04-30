# Sandwich Sync CLI

A command-line tool for fetching transactions from Rogers Bank API storing them locally and syncing them to LunchMoney.

## Features

- Parse curl-like commands to fetch transactions
- Store transactions in a SQLite database
- Interactive REPL interface
- List stored transactions
- Reconciliation and upload to LunchMoney

## Installation

1. Clone the repository
2. Build the application:

   ```
   go build -o lunchmoney
   ```
3. Rename `config.example.yaml` to `config.yaml` and add your key

## Usage

### Start the REPL

```
./lunchmoney repl
```

### REPL Commands

- `help` - Show help message
- `list` - List all transactions in the database
- `exit` or `quit` - Exit the REPL
- Enter a curl-like command to fetch transactions

### Example Curl Command

The application supports both single-quoted and caret-quoted curl commands:

#### Single-quoted (standard curl format)

```
curl 'https://rbaccess.rogersbank.com/issuing/digital/account/111111111/customer/00000000/activity?cycleStartDate=2025-04-23' \
  -H 'Accept-Language: en-US,en;q=0.9,es;q=0.8,en-CA;q=0.7' \
  -H 'Connection: keep-alive' \
  -b 'X_DOM_session_guid=215o1231; language=en' \
  -H 'Referer: https://rbaccess.rogersbank.com/app/transactions' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36' \
  -H 'accept: application/json' \
  -H 'content-type: application/json'
```

#### Caret-quoted (Windows CMD format)

```
curl ^"https://rbaccess.rogersbank.com/issuing/digital/account/111111111/customer/00000000/activity?cycleStartDate=2025-04-23^" ^
  -H ^"Accept-Language: en-US,en;q=0.9,es;q=0.8,en-CA;q=0.7^" ^
  -H ^"Connection: keep-alive^" ^
  -b ^"X_DOM_session_guid=41241; language=en^" ^
  -H ^"Referer: https://rbaccess.rogersbank.com/app/transactions^" ^
  -H ^"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36^" ^
  -H ^"accept: application/json^" ^
  -H ^"content-type: application/json^"
```

## Database

Transactions are stored in a SQLite database located at `~/.lunchmoney/transactions.db` by default. You can specify a different location using the `--db` flag:

```
./lunchmoney --db /path/to/database.db repl
```

## Testing

The project includes a comprehensive test suite. To run the tests:

```
go test ./...
```

This will run all tests in the project. You can also run tests for a specific package:

```
go test ./pkg/models
go test ./pkg/http
go test ./pkg/services
go test ./db
```

## Continuous Integration

This project uses GitHub Actions for continuous integration. The CI pipeline runs on every push to the main branch and on every pull request. It performs the following tasks:

1. Runs all tests
2. Lints the code using golangci-lint

The CI configuration is defined in `.github/workflows/go.yml`.

## Development

### Adding New Tests

When adding new features, please also add corresponding tests. The project follows a standard Go testing approach:

1. Create a file named `<filename>_test.go` in the same package as the code being tested
2. Write test functions with the naming convention `TestXxx` where `Xxx` is the function being tested
3. Use the `testing` package to write assertions

### Mocking

For testing components that have external dependencies, the project uses interface-based mocking:

- `db/mock_db.go` provides a mock implementation of the database
- `pkg/http/mock_lunchmoney.go` provides a mock implementation of the LunchMoney client