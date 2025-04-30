package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
)

type TransactionWithInstitution struct {
	Transaction
	Institution *Institution `json:"institution"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ReferenceNumber        string    `json:"referenceNumber"`
	ActivityType           string    `json:"activityType"`
	Amount                 *Amount   `json:"amount"`
	ActivityStatus         string    `json:"activityStatus"`
	ActivityCategory       string    `json:"activityCategory"`
	ActivityClassification string    `json:"activityClassification"`
	CardNumber             string    `json:"cardNumber"`
	Merchant               *Merchant `json:"merchant"`
	Date                   string    `json:"date"`
	ActivityCategoryCode   string    `json:"activityCategoryCode"`
	CustomerID             string    `json:"customerId"`
	PostedDate             string    `json:"postedDate"`
	LunchMoneyID           int64     `json:"lunchMoneyId"`
	Name                   *Name     `json:"name"`
}

// PrintFormatted prints the transaction in a formatted way
func (t *Transaction) PrintFormatted() {
	fmt.Printf("Transaction Details:\n")
	if t.ReferenceNumber != "" {
		fmt.Printf("	Reference Number: %s\n", t.ReferenceNumber)
	}
	if t.ActivityType != "" {
		fmt.Printf("	Activity Type: %s\n", t.ActivityType)
	}
	if t.Amount != nil && t.Amount.Value != "" && t.Amount.Currency != "" {
		fmt.Printf("	Amount: %s %s\n", t.Amount.Value, t.Amount.Currency)
	}
	if t.ActivityStatus != "" {
		fmt.Printf("	Activity Status: %s\n", t.ActivityStatus)
	}
	if t.ActivityCategory != "" {
		fmt.Printf("	Activity Category: %s\n", t.ActivityCategory)
	}
	if t.ActivityClassification != "" {
		fmt.Printf("	Activity Classification: %s\n", t.ActivityClassification)
	}
	if t.CardNumber != "" {
		fmt.Printf("	Card Number: %s\n", t.CardNumber)
	}
	if t.Merchant != nil {
		if t.Merchant.Name != "" {
			fmt.Printf("	Merchant Name: %s\n", t.Merchant.Name)
		}
		if t.Merchant.Category != "" {
			fmt.Printf("	Merchant Category: %s\n", t.Merchant.Category)
		}
		if t.Merchant.Address != nil {
			address := t.Merchant.Address
			if address.City != "" || address.StateProvince != "" || address.PostalCode != "" || address.CountryCode != "" {
				fmt.Printf("	Merchant Address: %s, %s, %s, %s\n",
					address.City,
					address.StateProvince,
					address.PostalCode,
					address.CountryCode)
			}
		}
	}
	if t.Date != "" {
		fmt.Printf("	Date: %s\n", t.Date)
	}
	if t.ActivityCategoryCode != "" {
		fmt.Printf("	Activity Category Code: %s\n", t.ActivityCategoryCode)
	}
	if t.CustomerID != "" {
		fmt.Printf("	Customer ID: %s\n", t.CustomerID)
	}
	if t.PostedDate != "" {
		fmt.Printf("	Posted Date: %s\n", t.PostedDate)
	}
	if t.LunchMoneyID != 0 {
		fmt.Printf("	LunchMoney ID: %d\n", t.LunchMoneyID)
	}
	if t.Name != nil && t.Name.NameOnCard != "" {
		fmt.Printf("	Name on Card: %s\n", t.Name.NameOnCard)
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
	Name                string   `json:"name"`
	CategoryCode        string   `json:"categoryCode"`
	CategoryDescription string   `json:"categoryDescription"`
	Category            string   `json:"category"`
	Address             *Address `json:"address"`
}

// Address represents a merchant's address
type Address struct {
	City          string `json:"city"`
	StateProvince string `json:"stateProvince"`
	PostalCode    string `json:"postalCode"`
	CountryCode   string `json:"countryCode"`
}

// Name represents the name on the card
type Name struct {
	NameOnCard string `json:"nameOnCard"`
}
