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
