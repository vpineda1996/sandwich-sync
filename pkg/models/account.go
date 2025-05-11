package models

import "time"

type AccountMapping struct {
	LunchMoneyId int64
	// External Name is optional
	ExternalName string
	IsPlaid      bool
}

type LunchMoneyAccount struct {
	// LunchMoneyId is the ID of the account in LunchMoney
	LunchMoneyId int64
	// Name is the name of the account in LunchMoney
	Name string
	// Balance is the balance of the account in LunchMoney
	Balance Amount
	// BalanceLastUpdated is the last time the balance was updated
	BalanceLastUpdated *time.Time
	// IsPlaid indicates if the account is linked via Plaid
	IsPlaid bool
}
