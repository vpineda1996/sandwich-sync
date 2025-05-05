package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/vpnda/sandwich-sync/pkg/http/lm"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

type InstitutionSelector struct {
	client          lm.LunchMoneyClientInterface
	selectedAccount *models.Institution
}

func NewInstitutionSelector(ctx context.Context, apiKey string) (*InstitutionSelector, error) {
	c, err := lm.NewLunchMoneyClient(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	return &InstitutionSelector{
		client: c,
	}, nil
}

// NewInstitutionSelectorWithClient creates a new institution selector with a provided client
func NewInstitutionSelectorWithClient(client lm.LunchMoneyClientInterface) *InstitutionSelector {
	return &InstitutionSelector{
		client: client,
	}
}

func (is *InstitutionSelector) FindPossibleInstitutionForTransaction(ctx context.Context,
	transaction *models.Transaction) (*models.Institution, error) {
	// Fetch institutions from the LunchMoney API
	institutions, err := is.client.ListInstitutions(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the transaction's merchant name matches any institution's card name
	someCardDigits := regexp.MustCompile("[0-9]{4}").FindString(transaction.Merchant.Name)
	if someCardDigits != "" {
		for _, institution := range institutions {
			if strings.Contains(institution.Name, someCardDigits) {
				return &institution, nil
			}
		}
	}

	if is.selectedAccount != nil {
		return is.selectedAccount, nil
	}

	fmt.Printf("Could not find institution for transaction [%s] %s (%s). Please select one:\n",
		transaction.ReferenceNumber, transaction.Merchant.Name, transaction.Amount.ToMoney().Display())
	return is.selectInstitutionInteractive()
}

func (is *InstitutionSelector) selectInstitutionInteractive() (*models.Institution, error) {
	institutions, err := is.client.ListInstitutions(context.Background())
	if err != nil {
		return nil, err
	}

	for i, institution := range institutions {
		fmt.Println("\t", i, institution.Name)
	}

	var selection int
	fmt.Print("Enter the number of the institution you want to select: ")
	_, err = fmt.Scan(&selection)
	if err != nil || selection < 0 || selection >= len(institutions) {
		return nil, fmt.Errorf("invalid selection")
	}

	return &institutions[selection], nil
}

func (is *InstitutionSelector) SelectDefaultInstitution() error {
	sa, err := is.selectInstitutionInteractive()
	if err != nil {
		return err
	}
	is.selectedAccount = sa
	return nil
}
