package webirr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	testBaseURL = "https://api.webirr.net"
	prodBaseURL = "https://api.webirr.net:8080"
)

// Client calls WeBirr merchant APIs.
type Client struct {
	merchantID string
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient uses a caller-owned reusable HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL overrides the gateway base URL. It is useful for tests and local gateways.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if strings.TrimSpace(baseURL) != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// NewClient creates a WeBirr client configured with merchant ID, API key, and environment.
func NewClient(merchantID, apiKey string, isTestEnv bool, options ...Option) *Client {
	baseURL := prodBaseURL
	if isTestEnv {
		baseURL = testBaseURL
	}
	client := &Client{
		merchantID: strings.TrimSpace(merchantID),
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
	for _, option := range options {
		if option != nil {
			option(client)
		}
	}
	return client
}

// CreateBill creates a new bill and returns the WeBirr payment code in ApiResponse.Res.
func (c *Client) CreateBill(ctx context.Context, bill *Bill) (*ApiResponse[string], error) {
	c.prepareBill(bill)
	return sendJSON[string](ctx, c, http.MethodPost, "einvoice/api/bill", nil, bill)
}

// UpdateBill updates an unpaid bill and returns OK in ApiResponse.Res.
func (c *Client) UpdateBill(ctx context.Context, bill *Bill) (*ApiResponse[string], error) {
	c.prepareBill(bill)
	return sendJSON[string](ctx, c, http.MethodPut, "einvoice/api/bill", nil, bill)
}

// DeleteBill deletes an unpaid bill by payment code.
func (c *Client) DeleteBill(ctx context.Context, paymentCode string) (*ApiResponse[string], error) {
	return send[string](ctx, c, http.MethodDelete, "einvoice/api/bill", map[string]string{"wbc_code": paymentCode}, nil)
}

// GetPaymentStatus gets the single-payment status for a payment code.
func (c *Client) GetPaymentStatus(ctx context.Context, paymentCode string) (*ApiResponse[PaymentStatus], error) {
	return send[PaymentStatus](ctx, c, http.MethodGet, "einvoice/api/paymentStatus", map[string]string{"wbc_code": paymentCode}, nil)
}

// GetBillByReference gets one bill by merchant bill reference.
func (c *Client) GetBillByReference(ctx context.Context, billReference string) (*ApiResponse[BillResponse], error) {
	return send[BillResponse](ctx, c, http.MethodGet, "einvoice/api/bill", map[string]string{"bill_reference": billReference}, nil)
}

// GetBillByPaymentCode gets one bill by WeBirr payment code.
func (c *Client) GetBillByPaymentCode(ctx context.Context, paymentCode string) (*ApiResponse[BillResponse], error) {
	return send[BillResponse](ctx, c, http.MethodGet, "einvoice/api/bill", map[string]string{"wbc_code": paymentCode}, nil)
}

// GetBills lists bills updated after a timestamp cursor.
func (c *Client) GetBills(ctx context.Context, paymentStatus int, lastTimeStamp string, limit int) (*ApiResponse[[]BillResponse], error) {
	return send[[]BillResponse](ctx, c, http.MethodGet, "einvoice/api/bills", map[string]string{
		"payment_status": strconv.Itoa(paymentStatus),
		"last_timestamp": lastTimeStamp,
		"limit":          strconv.Itoa(limit),
	}, nil)
}

// GetPayments lists payments updated after a timestamp cursor.
func (c *Client) GetPayments(ctx context.Context, lastTimeStamp string, limit int) (*ApiResponse[[]PaymentResponse], error) {
	return send[[]PaymentResponse](ctx, c, http.MethodGet, "einvoice/api/payments", map[string]string{
		"last_timestamp": lastTimeStamp,
		"limit":          strconv.Itoa(limit),
	}, nil)
}

// GetSupportedBanks gets banks and wallets configured for this merchant.
func (c *Client) GetSupportedBanks(ctx context.Context) (*ApiResponse[[]SupportedBank], error) {
	return send[[]SupportedBank](ctx, c, http.MethodGet, "einvoice/api/banks", nil, nil)
}

// GetStat retrieves basic merchant statistics for a date range.
func (c *Client) GetStat(ctx context.Context, dateFrom, dateTo string) (*ApiResponse[Stat], error) {
	return send[Stat](ctx, c, http.MethodGet, "merchant/stat", map[string]string{
		"date_from": dateFrom,
		"date_to":   dateTo,
	}, nil)
}

func (c *Client) prepareBill(bill *Bill) {
	if bill == nil {
		return
	}
	if c.merchantID != "" {
		bill.MerchantID = c.merchantID
	}
}

func sendJSON[T any](ctx context.Context, c *Client, method, path string, params map[string]string, body any) (*ApiResponse[T], error) {
	var requestBody io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(payload)
	}
	return send[T](ctx, c, method, path, params, requestBody)
}

func send[T any](ctx context.Context, c *Client, method, path string, params map[string]string, body io.Reader) (*ApiResponse[T], error) {
	request, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, params), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return &ApiResponse[T]{Error: fmt.Sprintf("http error %d %s", response.StatusCode, response.Status)}, nil
	}
	if len(strings.TrimSpace(string(responseBody))) == 0 {
		return &ApiResponse[T]{Error: "empty response"}, nil
	}

	var apiResponse ApiResponse[T]
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, err
	}
	return &apiResponse, nil
}

func (c *Client) buildURL(path string, params map[string]string) string {
	query := url.Values{}
	query.Set("api_key", c.apiKey)
	if c.merchantID != "" {
		query.Set("merchant_id", c.merchantID)
	}
	for key, value := range params {
		query.Set(key, value)
	}
	return fmt.Sprintf("%s/%s?%s", strings.TrimRight(c.baseURL, "/"), strings.TrimLeft(path, "/"), query.Encode())
}
