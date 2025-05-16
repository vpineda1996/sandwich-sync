package scotia

import (
	"fmt"

	"github.com/vpnda/sandwich-sync/pkg/models"
	"github.com/vpnda/sandwich-sync/pkg/utils"
	openapiclient "github.com/vpnda/scotiafetch"
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

func formatTransactionDescription(transaction *openapiclient.ApiCreditCreditIdTransactionsGet200ResponseDataSettledInner) string {
	switch TransactionType(transaction.GetTransactionType()) {
	case TransactionTypeDebit:
		// Only use merchant name on "purchases" / debit transactions
		return utils.Capitalize(*transaction.Merchant.Name)
	}
	return utils.Capitalize(transaction.GetCleanDescription())
}
