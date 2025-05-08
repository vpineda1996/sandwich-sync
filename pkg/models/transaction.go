package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
)

type TransactionWithAccountMapping struct {
	Transaction
	Mapping *AccountMapping `json:"accountMapping"`
}

type TransactionWithAccount struct {
	Transaction
	SourceAccountName string `json:"sourceAccountName"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ReferenceNumber string    `json:"referenceNumber"`
	LunchMoneyID    int64     `json:"lunchMoneyId"`
	Amount          Amount    `json:"amount"`
	Merchant        *Merchant `json:"merchant"`
	Date            string    `json:"date"`
	PostedDate      string    `json:"postedDate"`
}

// PrintFormatted prints the transaction in a formatted way
func (t *Transaction) PrintFormatted() {
	fmt.Printf("Transaction Details:\n")
	if t.ReferenceNumber != "" {
		fmt.Printf("	Reference Number: %s\n", t.ReferenceNumber)
	}
	if t.Amount.Value != "" && t.Amount.Currency != "" {
		fmt.Printf("	Amount: %s %s\n", t.Amount.Value, t.Amount.Currency)
	}

	if t.Merchant != nil {
		if t.Merchant.Name != "" {
			fmt.Printf("	Merchant Name: %s\n", t.Merchant.Name)
		}
		if t.Merchant.CategoryCode != "" {
			fmt.Printf("	Merchant Category: %s\n", t.Merchant.CategoryCode)
		}
		if t.Merchant.Address != nil {
			address := t.Merchant.Address
			if address.City != "" || address.StateProvince != "" {
				fmt.Printf("	Merchant Address: %s, %sn",
					address.City,
					address.StateProvince)
			}
		}
	}
	if t.Date != "" {
		fmt.Printf("	Date: %s\n", t.Date)
	}
	if t.PostedDate != "" {
		fmt.Printf("	Posted Date: %s\n", t.PostedDate)
	}
	if t.LunchMoneyID != 0 {
		fmt.Printf("	LunchMoney ID: %d\n", t.LunchMoneyID)
	}
}

// Amount represents a monetary amount
type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

func (a *Amount) ToMoney() *money.Money {
	split := strings.Split(a.Value, ".")
	currency := money.GetCurrency(a.Currency)
	if len(split) == 1 {
		split = append(split, "00")
	} else if len(split) == 2 && len(split[1]) < currency.Fraction {
		for i := 0; i < currency.Fraction-len(split[1]); i++ {
			split[1] += "0"
		}
	} else if len(split) == 2 && len(split[1]) >= currency.Fraction {
		split[1] = split[1][:currency.Fraction]
	}
	intTranslation, err := strconv.ParseInt(strings.Join(split, ""), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse amount: original split %v: %v", split, err))
	}
	return money.New(intTranslation, a.Currency)
}

// Merchant represents a merchant in a transaction
type Merchant struct {
	Name         string   `json:"name"`
	CategoryCode string   `json:"categoryCode"`
	Address      *Address `json:"address"`
}

// Address represents a merchant's address
type Address struct {
	City          string `json:"city"`
	StateProvince string `json:"stateProvince"`
}

// Name represents the name on the card
type Name struct {
	NameOnCard string `json:"nameOnCard"`
}
