package rogers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	"github.com/samber/lo"
	"github.com/vpnda/sandwich-sync/pkg/models"
)

type RogersBankClient struct {
	client      *http.Client
	fingerprint string

	accountId  string
	customerId string
}

func NewRogersBankClient(fingerprint string) *RogersBankClient {
	cookies, _ := cookiejar.New(nil)
	return &RogersBankClient{
		client: &http.Client{
			Jar: cookies,
		},
		fingerprint: fingerprint,
	}
}

const (
	rogersBaseURL = "https://rbaccess.rogersbank.com"

	localePath   = rogersBaseURL + "/issuing/digital/content/locale"
	authPath     = rogersBaseURL + "/issuing/digital/authenticate/user"
	activityPath = rogersBaseURL + "/issuing/digital/account/%s/customer/%s/activity"
)

type deviceInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type authRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	DeviceID   string `json:"deviceId"`
	DeviceInfo string `json:"deviceInfo"`
}

type authResponse struct {
	UserName string `json:"userName"`
	Accounts []struct {
		AccountID string `json:"accountId"`
		Customer  struct {
			CustomerID string `json:"customerId"`
		} `json:"customer"`
	} `json:"accounts"`
	Authenticated bool `json:"authenticated"`
}

func (c *RogersBankClient) Authenticate(ctx context.Context, username, password string) error {
	if c.IsAuthenticated() {
		return nil
	}

	// Step 1: GET /issuing/digital/content/locale to get SESSION cookie
	localeReq, _ := http.NewRequestWithContext(ctx,
		http.MethodGet, localePath, nil)
	localeReq.Header = getCommonHeaders()
	resp, err := c.client.Do(localeReq)
	if err != nil {
		return fmt.Errorf("failed to fetch locale content: %w", err)
	}
	resp.Body.Close()

	dvInfo := []deviceInfo{
		{"language", "en-US"},
		{"color_depth", "24"},
		{"java_enabled", "false"},
		{"browser_tz", "-240"},
		{"browser", "Chrome 135"},
		{"os", "Windows"},
	}

	deviceInfo := string(lo.Must(json.Marshal(dvInfo)))

	// Step 2: Authenticate using the SESSION cookie
	authPayload := authRequest{
		Username:   username,
		Password:   password, // Replace with real password
		DeviceID:   c.fingerprint,
		DeviceInfo: deviceInfo,
	}

	authBody, _ := json.Marshal(authPayload)
	authReq, _ := http.NewRequestWithContext(ctx,
		http.MethodPost, authPath, bytes.NewReader(authBody))
	authReq.Header = getCommonHeaders()
	authResp, err := c.client.Do(authReq)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	defer authResp.Body.Close()

	body, _ := io.ReadAll(authResp.Body)
	var authResult authResponse
	if err := json.Unmarshal(body, &authResult); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !authResult.Authenticated {
		return fmt.Errorf("authentication failed: %s", string(body))
	}

	if len(authResult.Accounts) == 0 {
		return fmt.Errorf("no accounts returned in auth response")
	}

	c.accountId = authResult.Accounts[0].AccountID
	c.customerId = authResult.Accounts[0].Customer.CustomerID

	fmt.Println("Account ID:", c.accountId)
	fmt.Println("Customer ID:", c.customerId)

	return nil
}

func (c *RogersBankClient) IsAuthenticated() bool {
	return c.accountId != "" && c.customerId != ""
}

func (c *RogersBankClient) FetchTransactions(ctx context.Context) ([]models.TransactionWithAccount, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("client is not authenticated")
	}

	activityReq, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(activityPath, c.accountId, c.customerId), nil)

	activityReq.Header = getCommonHeaders()
	activityResp, err := c.client.Do(activityReq)
	if err != nil {
		return nil, fmt.Errorf("activity request failed: %w", err)
	}
	defer activityResp.Body.Close()

	body, err := io.ReadAll(activityResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var transactions activitiesResponse
	if err := json.Unmarshal(body, &transactions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	var result []models.TransactionWithAccount
	for _, activity := range transactions.Activities {
		tx := models.TransactionWithAccount{
			Transaction:       activity,
			SourceAccountName: "Rogers Bank",
		}
		result = append(result, tx)
	}

	return result, nil
}

func getCommonHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0")
	headers.Set("Origin", "https://rbaccess.rogersbank.com")
	headers.Set("Referer", "https://rbaccess.rogersbank.com/?product=ROGERSBRAND&locale=en_CA")
	headers.Set("brand_id", "ROGERSBRAND")
	headers.Set("brand_locale", "en_CA")
	headers.Set("sourceType", "web")
	headers.Set("datatype", "json")
	return headers
}
