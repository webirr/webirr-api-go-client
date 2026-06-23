package webirr

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	createdAmount = "270.90"
	updatedAmount = "278.00"
	customerCode  = "sdk-test-customer"
	customerName  = "SDK Test Customer"
	customerPhone = "0911000000"
	description   = "SDK Test Bill"
	exampleCursor = "20251231"
)

func TestCreateBillSetsClientMerchantIDBeforeSending(t *testing.T) {
	bill := sampleBill("go/unit/1")
	client := NewClient("merchant-from-client", "x", true, WithHTTPClient(testHTTPClient(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(`{"error":null,"res":"123 456 789"}`), nil
	})))

	_, err := client.CreateBill(context.Background(), bill)
	if err != nil {
		t.Fatal(err)
	}
	if bill.MerchantID != "merchant-from-client" {
		t.Fatalf("merchantID = %q", bill.MerchantID)
	}
}

func TestPrepareBillDoesNotOverwriteExistingMerchantIDWhenClientMerchantIDIsEmpty(t *testing.T) {
	bill := sampleBill("go/unit/1")
	bill.MerchantID = "merchant-on-bill"
	client := NewClient("", "x", true)

	client.prepareBill(bill)

	if bill.MerchantID != "merchant-on-bill" {
		t.Fatalf("merchantID = %q", bill.MerchantID)
	}
}

func TestBuildURLIncludesMerchantIDForAllEndpointParameterShapesWhenConfigured(t *testing.T) {
	client := NewClient("merchant-from-client", "x", true)
	for name, endpoint := range endpointCases() {
		got := client.buildURL(endpoint.path, endpoint.params)
		if !strings.Contains(got, "merchant_id=merchant-from-client") {
			t.Fatalf("%s missing merchant_id: %s", name, got)
		}
	}
}

func TestBuildURLOmitsMerchantIDForAllEndpointParameterShapesWhenEmpty(t *testing.T) {
	client := NewClient("", "x", true)
	for name, endpoint := range endpointCases() {
		got := client.buildURL(endpoint.path, endpoint.params)
		if strings.Contains(got, "merchant_id=") {
			t.Fatalf("%s should omit merchant_id: %s", name, got)
		}
	}
}

func TestClientUsesInjectedHTTPClient(t *testing.T) {
	var requests []*http.Request
	client := NewClient("merchant-from-client", "x", true, WithHTTPClient(testHTTPClient(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req)
		return jsonResponse(`{"error":null,"res":"OK"}`), nil
	})))

	res, err := client.DeleteBill(context.Background(), "123 456 789")
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" || res.Res != "OK" {
		t.Fatalf("unexpected response: %+v", res)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d", len(requests))
	}
	rawURL := requests[0].URL.String()
	if requests[0].Method != http.MethodDelete {
		t.Fatalf("method = %s", requests[0].Method)
	}
	if !strings.Contains(rawURL, "merchant_id=merchant-from-client") {
		t.Fatalf("missing merchant_id: %s", rawURL)
	}
	if !strings.Contains(rawURL, "wbc_code=123+456+789") {
		t.Fatalf("missing encoded wbc_code: %s", rawURL)
	}
}

func TestTestEnvUsesDefaultDevBaseURL(t *testing.T) {
	t.Setenv("GATEWAY_URL", "")

	client := NewClient("merchant-from-client", "x", true)

	if client.baseURL != "https://api.webirr.dev" {
		t.Fatalf("baseURL = %q", client.baseURL)
	}
}

func TestTestEnvUsesGatewayURLEnvironmentOverride(t *testing.T) {
	t.Setenv("GATEWAY_URL", "https://local-gateway.example/")

	client := NewClient("merchant-from-client", "x", true)

	if client.baseURL != "https://local-gateway.example" {
		t.Fatalf("baseURL = %q", client.baseURL)
	}
}

func TestProductionIgnoresGatewayURLEnvironmentOverride(t *testing.T) {
	t.Setenv("GATEWAY_URL", "https://local-gateway.example/")

	client := NewClient("merchant-from-client", "x", false)

	if client.baseURL != "https://api.webirr.net:8080" {
		t.Fatalf("baseURL = %q", client.baseURL)
	}
}

func TestEndpointMethodsAndParameters(t *testing.T) {
	for name, tc := range endpointMethodCases() {
		t.Run(name, func(t *testing.T) {
			var request *http.Request
			client := NewClient("merchant-from-client", "api-key-from-client", true, WithHTTPClient(testHTTPClient(func(req *http.Request) (*http.Response, error) {
				request = req
				return jsonResponse(tc.responseBody), nil
			})))

			if err := tc.call(client); err != nil {
				t.Fatal(err)
			}
			if request == nil {
				t.Fatal("no request captured")
			}
			if request.Method != tc.method {
				t.Fatalf("method = %s", request.Method)
			}
			if request.URL.Path != tc.path {
				t.Fatalf("path = %s", request.URL.Path)
			}
			query := request.URL.Query()
			if query.Get("api_key") != "api-key-from-client" {
				t.Fatalf("api_key = %q", query.Get("api_key"))
			}
			if query.Get("merchant_id") != "merchant-from-client" {
				t.Fatalf("merchant_id = %q", query.Get("merchant_id"))
			}
			for key, value := range tc.query {
				if query.Get(key) != value {
					t.Fatalf("%s = %q", key, query.Get(key))
				}
			}
		})
	}
}

func TestAPIErrorResponsesReturnTypedResponseWithoutTransportError(t *testing.T) {
	for name, call := range apiErrorCalls() {
		t.Run(name, func(t *testing.T) {
			client := NewClient("merchant-from-client", "bad-key", true, WithHTTPClient(testHTTPClient(func(req *http.Request) (*http.Response, error) {
				return jsonResponse(`{"error":"invalid api key","errorCode":"ERROR_INVALID_API_KEY","res":null}`), nil
			})))

			errorMessage, errorCode, err := call(client)
			if err != nil {
				t.Fatal(err)
			}
			if errorMessage != "invalid api key" || errorCode != "ERROR_INVALID_API_KEY" {
				t.Fatalf("api error = %q/%q", errorMessage, errorCode)
			}
		})
	}
}

func TestHTTPStatusErrorReturnsAPIResponseWithoutTransportError(t *testing.T) {
	client := NewClient("merchant-from-client", "x", true, WithHTTPClient(testHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("gateway unavailable")),
		}, nil
	})))

	response, err := client.DeleteBill(context.Background(), "123 456 789")
	if err != nil {
		t.Fatal(err)
	}
	if response.Error != "http error 502 502 Bad Gateway" {
		t.Fatalf("error = %q", response.Error)
	}
}

func TestBillSerializesCustomerPhone(t *testing.T) {
	payload := marshalBill(t, sampleBill("go/unit/1"))
	if payload["customerPhone"] != customerPhone {
		t.Fatalf("customerPhone = %v", payload["customerPhone"])
	}
}

func TestBillSerializesWithoutCustomerPhoneAsEmptyString(t *testing.T) {
	bill := sampleBill("go/unit/1")
	bill.CustomerPhone = ""
	payload := marshalBill(t, bill)
	if payload["customerPhone"] != "" {
		t.Fatalf("customerPhone = %v", payload["customerPhone"])
	}
}

func TestBillSerializesEmptyExtrasAsJSONObject(t *testing.T) {
	payload := marshalBill(t, sampleBill("go/unit/1"))
	extras, ok := payload["extras"].(map[string]any)
	if !ok {
		t.Fatalf("extras type = %T", payload["extras"])
	}
	if len(extras) != 0 {
		t.Fatalf("extras = %#v", extras)
	}
}

func TestBillSerializesPopulatedExtrasAsJSONObject(t *testing.T) {
	bill := sampleBill("go/unit/1")
	bill.Extras = map[string]string{"source": "unit-test"}
	payload := marshalBill(t, bill)
	extras, ok := payload["extras"].(map[string]any)
	if !ok {
		t.Fatalf("extras type = %T", payload["extras"])
	}
	if extras["source"] != "unit-test" {
		t.Fatalf("source = %v", extras["source"])
	}
}

func TestBillResponseDeserializesRetrievalOnlyFields(t *testing.T) {
	var response ApiResponse[BillResponse]
	mustUnmarshal(t, `{
		"error": null,
		"res": {
			"customerCode": "SDK-TEST-CUSTOMER",
			"customerName": "SDK Test Customer",
			"customerPhone": "0911000000",
			"billReference": "go/unit/1",
			"time": "2026-06-12 10:00",
			"description": "SDK Test Bill",
			"amount": "278.00",
			"merchantID": "merchant-from-client",
			"wbcCode": "123 456 789",
			"paymentStatus": 0,
			"updateTimeStamp": "2026061210000000000"
		}
	}`, &response)

	if response.Error != "" {
		t.Fatalf("error = %q", response.Error)
	}
	if response.Res.CustomerPhone != customerPhone {
		t.Fatalf("customerPhone = %q", response.Res.CustomerPhone)
	}
	if response.Res.WbcCode != "123 456 789" {
		t.Fatalf("wbcCode = %q", response.Res.WbcCode)
	}
	if response.Res.UpdateTimeStamp != "2026061210000000000" {
		t.Fatalf("updateTimeStamp = %q", response.Res.UpdateTimeStamp)
	}
}

func TestPaymentResponseUsesPaymentDateAsTimeAlias(t *testing.T) {
	var response ApiResponse[[]PaymentResponse]
	mustUnmarshal(t, `{
		"error": null,
		"res": [{
			"status": 2,
			"id": 101,
			"bankID": "test-bank",
			"paymentReference": "TX-1",
			"paymentDate": "2026-06-12 10:11:12",
			"confirmed": true,
			"confirmedTime": "2026-06-12 10:12:12",
			"canceled": false,
			"canceledTime": "0001-01-01 00:00:00",
			"amount": "278.00",
			"wbcCode": "123 456 789",
			"updateTimeStamp": "2026061210121200000"
		}]
	}`, &response)

	payment := response.Res[0]
	if payment.PaymentDate != "2026-06-12 10:11:12" || payment.Time != payment.PaymentDate {
		t.Fatalf("payment date/time = %q/%q", payment.PaymentDate, payment.Time)
	}
	if !payment.IsPaid() {
		t.Fatal("expected paid payment")
	}
}

func TestPaymentDetailKeepsLegacyTimeAsPaymentDateAlias(t *testing.T) {
	var response ApiResponse[PaymentStatus]
	mustUnmarshal(t, `{
		"error": null,
		"res": {
			"status": 2,
			"data": {
				"status": 2,
				"id": 101,
				"bankID": "test-bank",
				"paymentReference": "TX-1",
				"time": "2026-06-12 10:11:12",
				"confirmed": true,
				"confirmedTime": "2026-06-12 10:12:12",
				"amount": "278.00",
				"wbcCode": "123 456 789",
				"updateTimeStamp": "2026061210121200000"
			}
		}
	}`, &response)

	if !response.Res.IsPaid() {
		t.Fatal("expected status paid")
	}
	if response.Res.Data == nil {
		t.Fatal("expected payment data")
	}
	if response.Res.Data.PaymentDate != "2026-06-12 10:11:12" || response.Res.Data.Time != response.Res.Data.PaymentDate {
		t.Fatalf("payment date/time = %q/%q", response.Res.Data.PaymentDate, response.Res.Data.Time)
	}
}

func TestStatDeserializesGatewayPascalCaseFields(t *testing.T) {
	var response ApiResponse[Stat]
	mustUnmarshal(t, `{
		"error": null,
		"res": {
			"NBills": 10,
			"NBillsPaid": 4,
			"NBillsUnpaid": 6,
			"AmountBills": "100.00",
			"AmountPaid": "40.00",
			"AmountUnpaid": "60.00"
		}
	}`, &response)

	if response.Res.NBills != 10 || response.Res.NBillsPaid != 4 || response.Res.NBillsUnpaid != 6 {
		t.Fatalf("unexpected stat counts: %+v", response.Res)
	}
	if response.Res.AmountBills != "100.00" || response.Res.AmountPaid != "40.00" || response.Res.AmountUnpaid != "60.00" {
		t.Fatalf("unexpected stat amounts: %+v", response.Res)
	}
}

func TestSupportedBankDeserializesGatewayFields(t *testing.T) {
	var response ApiResponse[[]SupportedBank]
	mustUnmarshal(t, `{
		"error": null,
		"res": [{
			"bankID": "cbe_mobile",
			"name": "CBE Mobile Banking"
		}]
	}`, &response)

	if response.Res[0].BankID != "cbe_mobile" || response.Res[0].Name != "CBE Mobile Banking" {
		t.Fatalf("bank = %+v", response.Res[0])
	}
}

func sampleBill(reference string) *Bill {
	return &Bill{
		Amount:        createdAmount,
		CustomerCode:  customerCode,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Time:          "2026-06-12 10:00",
		Description:   description,
		BillReference: reference,
	}
}

func marshalBill(t *testing.T, bill *Bill) map[string]any {
	t.Helper()
	payload, err := json.Marshal(bill)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	mustUnmarshal(t, string(payload), &result)
	return result
}

type endpointCase struct {
	path   string
	params map[string]string
}

func endpointCases() map[string]endpointCase {
	return map[string]endpointCase{
		"createBill":           {path: "einvoice/api/bill"},
		"updateBill":           {path: "einvoice/api/bill"},
		"deleteBill":           {path: "einvoice/api/bill", params: map[string]string{"wbc_code": "123 456 789"}},
		"getPaymentStatus":     {path: "einvoice/api/paymentStatus", params: map[string]string{"wbc_code": "123 456 789"}},
		"getBillByReference":   {path: "einvoice/api/bill", params: map[string]string{"bill_reference": "go/unit/1"}},
		"getBillByPaymentCode": {path: "einvoice/api/bill", params: map[string]string{"wbc_code": "123 456 789"}},
		"getBills":             {path: "einvoice/api/bills", params: map[string]string{"payment_status": "-1", "last_timestamp": exampleCursor, "limit": "10"}},
		"getPayments":          {path: "einvoice/api/payments", params: map[string]string{"last_timestamp": exampleCursor, "limit": "10"}},
		"getSupportedBanks":    {path: "einvoice/api/banks"},
		"getStat":              {path: "merchant/stat", params: map[string]string{"date_from": "2025-01-01", "date_to": "2025-01-02"}},
	}
}

type endpointMethodCase struct {
	method       string
	path         string
	query        map[string]string
	responseBody string
	call         func(*Client) error
}

func endpointMethodCases() map[string]endpointMethodCase {
	return map[string]endpointMethodCase{
		"createBill": {
			method:       http.MethodPost,
			path:         "/einvoice/api/bill",
			responseBody: `{"error":null,"res":"123 456 789"}`,
			call: func(client *Client) error {
				_, err := client.CreateBill(context.Background(), sampleBill("go/unit/create"))
				return err
			},
		},
		"updateBill": {
			method:       http.MethodPut,
			path:         "/einvoice/api/bill",
			responseBody: `{"error":null,"res":"OK"}`,
			call: func(client *Client) error {
				_, err := client.UpdateBill(context.Background(), sampleBill("go/unit/update"))
				return err
			},
		},
		"deleteBill": {
			method:       http.MethodDelete,
			path:         "/einvoice/api/bill",
			query:        map[string]string{"wbc_code": "123 456 789"},
			responseBody: `{"error":null,"res":"OK"}`,
			call: func(client *Client) error {
				_, err := client.DeleteBill(context.Background(), "123 456 789")
				return err
			},
		},
		"getPaymentStatus": {
			method:       http.MethodGet,
			path:         "/einvoice/api/paymentStatus",
			query:        map[string]string{"wbc_code": "123 456 789"},
			responseBody: `{"error":null,"res":{"status":0,"data":null}}`,
			call: func(client *Client) error {
				_, err := client.GetPaymentStatus(context.Background(), "123 456 789")
				return err
			},
		},
		"getBillByReference": {
			method:       http.MethodGet,
			path:         "/einvoice/api/bill",
			query:        map[string]string{"bill_reference": "go/unit/1"},
			responseBody: `{"error":null,"res":{"billReference":"go/unit/1","wbcCode":"123 456 789"}}`,
			call: func(client *Client) error {
				_, err := client.GetBillByReference(context.Background(), "go/unit/1")
				return err
			},
		},
		"getBillByPaymentCode": {
			method:       http.MethodGet,
			path:         "/einvoice/api/bill",
			query:        map[string]string{"wbc_code": "123 456 789"},
			responseBody: `{"error":null,"res":{"billReference":"go/unit/1","wbcCode":"123 456 789"}}`,
			call: func(client *Client) error {
				_, err := client.GetBillByPaymentCode(context.Background(), "123 456 789")
				return err
			},
		},
		"getBills": {
			method:       http.MethodGet,
			path:         "/einvoice/api/bills",
			query:        map[string]string{"payment_status": "-1", "last_timestamp": exampleCursor, "limit": "10"},
			responseBody: `{"error":null,"res":[]}`,
			call: func(client *Client) error {
				_, err := client.GetBills(context.Background(), -1, exampleCursor, 10)
				return err
			},
		},
		"getPayments": {
			method:       http.MethodGet,
			path:         "/einvoice/api/payments",
			query:        map[string]string{"last_timestamp": exampleCursor, "limit": "10"},
			responseBody: `{"error":null,"res":[]}`,
			call: func(client *Client) error {
				_, err := client.GetPayments(context.Background(), exampleCursor, 10)
				return err
			},
		},
		"getSupportedBanks": {
			method:       http.MethodGet,
			path:         "/einvoice/api/banks",
			responseBody: `{"error":null,"res":[]}`,
			call: func(client *Client) error {
				_, err := client.GetSupportedBanks(context.Background())
				return err
			},
		},
		"getStat": {
			method:       http.MethodGet,
			path:         "/merchant/stat",
			query:        map[string]string{"date_from": "2025-01-01", "date_to": "2025-01-02"},
			responseBody: `{"error":null,"res":{"NBills":0}}`,
			call: func(client *Client) error {
				_, err := client.GetStat(context.Background(), "2025-01-01", "2025-01-02")
				return err
			},
		},
	}
}

type apiErrorCall func(*Client) (string, string, error)

func apiErrorCalls() map[string]apiErrorCall {
	return map[string]apiErrorCall{
		"createBill": func(client *Client) (string, string, error) {
			res, err := client.CreateBill(context.Background(), sampleBill("go/unit/error-create"))
			return res.Error, res.ErrorCode, err
		},
		"updateBill": func(client *Client) (string, string, error) {
			res, err := client.UpdateBill(context.Background(), sampleBill("go/unit/error-update"))
			return res.Error, res.ErrorCode, err
		},
		"deleteBill": func(client *Client) (string, string, error) {
			res, err := client.DeleteBill(context.Background(), "123 456 789")
			return res.Error, res.ErrorCode, err
		},
		"getPaymentStatus": func(client *Client) (string, string, error) {
			res, err := client.GetPaymentStatus(context.Background(), "123 456 789")
			return res.Error, res.ErrorCode, err
		},
		"getBillByReference": func(client *Client) (string, string, error) {
			res, err := client.GetBillByReference(context.Background(), "go/unit/1")
			return res.Error, res.ErrorCode, err
		},
		"getBillByPaymentCode": func(client *Client) (string, string, error) {
			res, err := client.GetBillByPaymentCode(context.Background(), "123 456 789")
			return res.Error, res.ErrorCode, err
		},
		"getBills": func(client *Client) (string, string, error) {
			res, err := client.GetBills(context.Background(), -1, exampleCursor, 10)
			return res.Error, res.ErrorCode, err
		},
		"getPayments": func(client *Client) (string, string, error) {
			res, err := client.GetPayments(context.Background(), exampleCursor, 10)
			return res.Error, res.ErrorCode, err
		},
		"getSupportedBanks": func(client *Client) (string, string, error) {
			res, err := client.GetSupportedBanks(context.Background())
			return res.Error, res.ErrorCode, err
		},
		"getStat": func(client *Client) (string, string, error) {
			res, err := client.GetStat(context.Background(), "2025-01-01", "2025-01-02")
			return res.Error, res.ErrorCode, err
		},
	}
}

func mustUnmarshal(t *testing.T, input string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(input), target); err != nil {
		t.Fatal(err)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testHTTPClient(handler roundTripFunc) *http.Client {
	return &http.Client{Transport: handler}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
