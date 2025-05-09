package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/vpnda/sandwich-sync/pkg/http/scotia"
)

func main() {
	c := lo.Must(scotia.NewScotiaClient())
	lo.Must0(c.AuthenticateDynamic(context.Background()))
	transactions := lo.Must(c.FetchTransactions(context.Background()))

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
	// scotia.Interact()
}
