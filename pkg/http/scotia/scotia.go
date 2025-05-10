package scotia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/vpnda/sandwich-sync/pkg/models"
	openapiclient "github.com/vpnda/scotiafetch"
)

type ScotiaClient struct {
	authClient *http.Client
	apiClient  *openapiclient.APIClient
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

var (
	ErrAuthRedirect = fmt.Errorf("got redirect")
)

func NewScotiaClient() (*ScotiaClient, error) {
	configuration := openapiclient.NewConfiguration()

	jar, _ := cookiejar.New(nil)
	client := http.Client{
		Jar: jar,
		// Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// 	d, _ := httputil.DumpRequest(r, true)
		// 	fmt.Println(string(d))
		// 	res, err := http.DefaultTransport.RoundTrip(r)
		// 	if err == nil {
		// 		d, _ := httputil.DumpResponse(res, true)
		// 		fmt.Println(string(d))
		// 	}
		// 	return res, err
		// }),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Print the redirect URL
			log.Info().Msgf("request is being redirected: %s", req.URL.String())
			// You can also modify the request here if needed
			return ErrAuthRedirect // Returning nil means we follow the redirect
		},
	}
	configuration.HTTPClient = &client
	apiClient := openapiclient.NewAPIClient(configuration)

	return &ScotiaClient{
		authClient: &client,
		apiClient:  apiClient,
	}, nil
}

type AccountProductCategory string
type AccountType string
type TransactionType string

const (
	AccountCategoryDayToDay    AccountProductCategory = "DAYTODAY"
	AccountCategoryCreditCards AccountProductCategory = "CREDITCARDS"
	AccountCategoryInvesting   AccountProductCategory = "INVESTING"

	AccountTypeChequing AccountType = "Chequing"

	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

// FetchTransactions implements http.TransactionFetcher.
func (s *ScotiaClient) FetchTransactions(ctx context.Context) ([]models.TransactionWithAccount, error) {
	resp, r, err := s.apiClient.DefaultAPI.ApiAccountsSummaryGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts summary: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("no response from API")
	}

	var result []models.TransactionWithAccount
	depositAccounts := lo.Filter(resp.Data.GetProducts(), func(p openapiclient.ApiAccountsSummaryGet200ResponseDataProductsInner, _ int) bool {
		return *p.Type == string(AccountTypeChequing)
	})

	r.Body.Close()

	for _, account := range depositAccounts {
		key := account.GetKey()
		transactions, r, err := s.apiClient.DefaultAPI.ApiTransactionsDepositAccountsDepositAccountIdGet(ctx, key).Execute()

		if err != nil {
			return nil, fmt.Errorf("failed to fetch transactions for account %s: %w", key, err)
		}

		if transactions == nil {
			return nil, fmt.Errorf("no transactions found for account %s", key)
		}

		for _, transaction := range transactions.GetData() {
			// Convert to models.TransactionWithAccount
			transactionWithAccount := models.TransactionWithAccount{
				SourceAccountName: account.GetDescription(),
				Transaction: models.Transaction{
					ReferenceNumber: *transaction.Key,
					Amount: formatAmount(TransactionType(transaction.GetTransactionType()),
						transaction.GetTransactionAmount()),
					Merchant: &models.Merchant{
						Name: *transaction.CleanDescription,
					},
					Date: *transaction.TransactionDate,
				},
			}
			result = append(result, transactionWithAccount)
		}
		r.Body.Close()
	}

	creditCardAccounts := lo.Filter(resp.Data.GetProducts(), func(p openapiclient.ApiAccountsSummaryGet200ResponseDataProductsInner, _ int) bool {
		return *p.ProductCategory == string(AccountCategoryCreditCards)
	})

	for _, account := range creditCardAccounts {
		key := account.GetKey()
		transactions, r, err := s.apiClient.DefaultAPI.ApiCreditCreditIdTransactionsGet(ctx, key).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch transactions for account %s: %w", key, err)
		}
		if transactions == nil {
			return nil, fmt.Errorf("no transactions found for account %s", key)
		}

		for _, transaction := range transactions.GetData().Settled {
			date, err := time.Parse("2006-01-02T15:04:05", *transaction.TransactionDate)
			if err != nil {
				return nil, fmt.Errorf("failed to parse transaction date %s: %w", *transaction.TransactionDate, err)
			}
			transactionWithAccount := models.TransactionWithAccount{
				SourceAccountName: account.GetDescription(),
				Transaction: models.Transaction{
					ReferenceNumber: *transaction.Key,
					Amount: formatAmount(
						TransactionType(*transaction.TransactionType),
						transaction.GetTransactionAmount()),
					Merchant: &models.Merchant{
						Name: *transaction.CleanDescription,
					},
					Date: date.Format(time.DateOnly),
				},
			}
			result = append(result, transactionWithAccount)
		}
		r.Body.Close()
	}

	return result, nil
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

func (s *ScotiaClient) validateHealthySession(ctx context.Context, headers map[string]string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://secure.scotiabank.com/api/accounts/summary", nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	res, err := s.authClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("request failed with status %d: %s", res.StatusCode, string(body))
	}

	err = json.NewDecoder(res.Body).Decode(lo.ToPtr(map[string]any{}))
	if err != nil {
		return err
	}

	return nil

}
