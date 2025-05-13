package scotia

import (
	"fmt"
	"strings"

	"github.com/vpnda/sandwich-sync/pkg/models"
	openapiclient "github.com/vpnda/scotiafetch"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func AccountName(account interface{ GetDescription() string }) string {
	return account.GetDescription()
}

func formatAmount(transactionType TransactionType,
	transactionAmount openapiclient.ApiAccountsSummaryGet200ResponseDataProductsInnerPrimaryBalancesInner) models.Amount {
	amountStr := fmt.Sprintf("%.2f", *transactionAmount.Amount)
	if transactionType == TransactionTypeCredit {
		// The database stores outflows as positive values, so wehn the transaction type is
		// credit, we need to negate the amount as an inflow.
		amountStr = "-" + amountStr
	}
	return models.Amount{
		Value:    amountStr,
		Currency: *transactionAmount.CurrencyCode,
	}
}

func capitalize(s string) string {
	return cases.Title(language.English).String(strings.ToLower(s))
}

func formatTransactionDescription(transaction *openapiclient.ApiCreditCreditIdTransactionsGet200ResponseDataSettledInner) string {
	switch TransactionType(transaction.GetTransactionType()) {
	case TransactionTypeDebit:
		// Only use merchant name on "purchases" / debit transactions
		return capitalize(*transaction.Merchant.Name)
	}
	return capitalize(transaction.GetCleanDescription())
}
