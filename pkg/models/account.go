package models

import "time"

type AccountMapping struct {
	LunchMoneyId int64
	// External Name is optional
	ExternalName string
	IsPlaid      bool
}

type ExternalAccount struct {
	// Name is a unique name for the account
	Name string
	// Description is a human readable description of the account
	// it is optional and can be empty
	Description string
	// Balance is the balance of the account
	Balance Amount
}

type LunchMoneyAccount struct {
	// LunchMoneyId is the ID of the account in LunchMoney
	LunchMoneyId int64
	// Name is the name of the account in LunchMoney
	Name string
	// DisplayName is the display name of the account in LunchMoney
	// this can be empty
	DisplayName string
	// Balance is the balance of the account in LunchMoney
	Balance Amount
	// BalanceLastUpdated is the last time the balance was updated
	BalanceLastUpdated *time.Time
	// IsPlaid indicates if the account is linked via Plaid
	IsPlaid bool
	// ShouldSync indicates if the account should be synced
	ShouldSync bool
}
