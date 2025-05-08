package models

type AccountMapping struct {
	LunchMoneyId int64
	// External Name is optional
	ExternalName string
	IsPlaid      bool
}

type LunchMoneyAccount struct {
	LunchMoneyId int64
	Name         string
	IsPlaid      bool
}
