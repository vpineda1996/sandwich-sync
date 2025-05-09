# Sandwich Sync CLI

A command-line tool for fetching transactions from Rogers/Wealthsimple and Scotiabank API storing them locally and syncing them to LunchMoney.

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
4. If you are going to use Scotia, you'll need to install [`patchright`](https://github.com/Kaliiiiiiiiii-Vinyzu/patchright-python)(a modified version of [playwright](https://playwright.dev/)) to bypass Akamai logins.
```
# Install Patchright with Pip from PyPI
pip install patchright

# Install Chromium-Driver for Patchright
patchright install chromium
```

## Usage

### Start the REPL

```
./lunchmoney repl
```

### REPL Commands

- `help` - Show help message
- `list` - List all transactions in the database
- `exit` or `quit` - Exit the REPL
- `fetch <provider>` - Fetch recent transactions from a provider (rogers, wealthsimple, scotiabank)
- `sync` - Sync transactions to LunchMoney

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