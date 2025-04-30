package models

import (
	"testing"
)

func TestAmountToMoney(t *testing.T) {
	testCases := []struct {
		name           string
		amount         Amount
		expectedAmount int64
		expectedCurr   string
	}{
		{
			name:           "Whole number",
			amount:         Amount{Value: "100", Currency: "USD"},
			expectedAmount: 10000,
			expectedCurr:   "USD",
		},
		{
			name:           "Decimal number",
			amount:         Amount{Value: "25.99", Currency: "USD"},
			expectedAmount: 2599,
			expectedCurr:   "USD",
		},
		{
			name:           "Single decimal place",
			amount:         Amount{Value: "10.5", Currency: "USD"},
			expectedAmount: 1050,
			expectedCurr:   "USD",
		},
		{
			name:           "Different currency",
			amount:         Amount{Value: "50.75", Currency: "EUR"},
			expectedAmount: 5075,
			expectedCurr:   "EUR",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.amount.ToMoney()

			if result.Amount() != tc.expectedAmount {
				t.Errorf("Expected amount %d, got %d", tc.expectedAmount, result.Amount())
			}

			if result.Currency().Code != tc.expectedCurr {
				t.Errorf("Expected currency %s, got %s", tc.expectedCurr, result.Currency().Code)
			}
		})
	}
}

func TestTransactionPrintFormatted(t *testing.T) {
	// This is a visual test that's hard to verify programmatically
	// We'll just ensure it doesn't panic
	tx := &Transaction{
		ReferenceNumber: "TEST123",
		ActivityType:    "TRANS",
		Amount: &Amount{
			Value:    "25.99",
			Currency: "USD",
		},
		ActivityStatus:         "APPROVED",
		ActivityCategory:       "PURCHASE",
		ActivityClassification: "PURCHASE",
		CardNumber:             "************1234",
		Merchant: &Merchant{
			Name:     "Test Merchant",
			Category: "RETAIL",
			Address: &Address{
				City:          "Test City",
				StateProvince: "TS",
				PostalCode:    "12345",
				CountryCode:   "US",
			},
		},
		Date:                 "2025-04-29",
		ActivityCategoryCode: "0001",
		CustomerID:           "TEST",
		PostedDate:           "2025-04-29",
		Name:                 &Name{NameOnCard: "TEST USER"},
		LunchMoneyID:         12345,
	}

	// This should not panic
	tx.PrintFormatted()
}

func TestMoneyEquals(t *testing.T) {
	// Test money equality
	amount1 := &Amount{Value: "25.99", Currency: "USD"}
	amount2 := &Amount{Value: "25.99", Currency: "USD"}
	amount3 := &Amount{Value: "30.00", Currency: "USD"}
	amount4 := &Amount{Value: "25.99", Currency: "EUR"}

	money1 := amount1.ToMoney()
	money2 := amount2.ToMoney()
	money3 := amount3.ToMoney()
	money4 := amount4.ToMoney()

	// Test equality
	equal12, err := money1.Equals(money2)
	if err != nil {
		t.Fatalf("Error comparing money: %v", err)
	}
	if !equal12 {
		t.Errorf("Expected money1 and money2 to be equal")
	}

	// Test inequality with different amounts
	equal13, err := money1.Equals(money3)
	if err != nil {
		t.Fatalf("Error comparing money: %v", err)
	}
	if equal13 {
		t.Errorf("Expected money1 and money3 to be unequal")
	}

	// Test inequality with different currencies
	_, err = money1.Equals(money4)
	if err == nil {
		t.Errorf("Expected error when comparing different currencies")
	}
}
