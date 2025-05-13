package scotia

import "fmt"

type AccountProductCategory string
type AccountType string
type TransactionType string

// TODO - Add more account types and maybe move this to the actual client
const (
	AccountCategoryDayToDay    AccountProductCategory = "DAYTODAY"
	AccountCategoryCreditCards AccountProductCategory = "CREDITCARDS"
	AccountCategoryInvesting   AccountProductCategory = "INVESTING"

	AccountTypeChequing AccountType = "Chequing"

	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

var (
	ErrAuthRedirect      = fmt.Errorf("got redirect")
	ErrReadingConfigFile = fmt.Errorf("error reading config file")
)
