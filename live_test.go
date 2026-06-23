package webirr

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var livePaymentCode string
var liveBillReference string
var liveBillUpdateTimeStamp string
var liveDeleted bool

func TestLiveTestEnvCreateUpdateStatusLookupListPaymentsStatBanksDelete(t *testing.T) {
	client := liveTestEnvClient(t)
	ctx := context.Background()
	liveBillReference = "go/test/" + time.Now().UTC().Format("20060102150405")

	createBill := sampleBill(liveBillReference)
	createBill.Time = time.Now().Format("2006-01-02 15:04")
	create, err := client.CreateBill(ctx, createBill)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, create)
	if createBill.MerchantID == "" {
		t.Fatal("client did not set bill merchantID")
	}
	livePaymentCode = create.Res
	t.Cleanup(func() {
		if !liveDeleted && livePaymentCode != "" {
			_, _ = client.DeleteBill(context.Background(), livePaymentCode)
		}
	})
	if livePaymentCode == "" || !looksLikePaymentCode(livePaymentCode) {
		t.Fatalf("unexpected payment code: %q", livePaymentCode)
	}

	updateBill := sampleBill(liveBillReference)
	updateBill.Amount = updatedAmount
	updateBill.CustomerName = "SDK Test Customer Updated"
	updateBill.Time = createBill.Time
	update, err := client.UpdateBill(ctx, updateBill)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, update)
	if strings.ToLower(update.Res) != "ok" {
		t.Fatalf("update res = %q", update.Res)
	}

	status, err := client.GetPaymentStatus(ctx, livePaymentCode)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, status)
	if status.Res.Status != 0 {
		t.Fatalf("new bill status = %d", status.Res.Status)
	}

	byReference, err := client.GetBillByReference(ctx, liveBillReference)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, byReference)
	assertLiveBill(t, byReference.Res, liveBillReference, livePaymentCode)
	liveBillUpdateTimeStamp = byReference.Res.UpdateTimeStamp

	byCode, err := client.GetBillByPaymentCode(ctx, livePaymentCode)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, byCode)
	assertLiveBill(t, byCode.Res, liveBillReference, livePaymentCode)

	bills, err := client.GetBills(ctx, 0, "20251231", 100)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, bills)
	if bills.Res == nil {
		t.Fatal("expected bills slice")
	}

	payments, err := client.GetPayments(ctx, "20251231", 10)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, payments)
	if payments.Res == nil {
		t.Fatal("expected payments slice")
	}

	stat, err := client.GetStat(ctx, "2025-01-01", "2030-01-31")
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, stat)

	banks, err := client.GetSupportedBanks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, banks)
	if len(banks.Res) == 0 {
		t.Fatal("expected merchant-supported banks")
	}
	for _, bank := range banks.Res {
		if bank.BankID == "" || bank.Name == "" {
			t.Fatalf("invalid bank: %+v", bank)
		}
	}

	deleteResponse, err := client.DeleteBill(ctx, livePaymentCode)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAPIError(t, deleteResponse)
	if strings.ToLower(deleteResponse.Res) != "ok" {
		t.Fatalf("delete res = %q", deleteResponse.Res)
	}
	liveDeleted = true
}

func liveTestEnvClient(t *testing.T) *Client {
	t.Helper()
	merchantID := os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID")
	apiKey := os.Getenv("WEBIRR_TEST_ENV_API_KEY")
	if merchantID == "" || apiKey == "" {
		t.Skip("WEBIRR_TEST_ENV_MERCHANT_ID and WEBIRR_TEST_ENV_API_KEY are required for TestEnv smoke tests")
	}
	return NewClient(merchantID, apiKey, true)
}

func assertNoAPIError[T any](t *testing.T, response *ApiResponse[T]) {
	t.Helper()
	if response == nil {
		t.Fatal("nil response")
	}
	if response.Error != "" || response.ErrorCode != "" {
		t.Fatalf("api error: error=%q errorCode=%q", response.Error, response.ErrorCode)
	}
}

func assertLiveBill(t *testing.T, bill BillResponse, reference, paymentCode string) {
	t.Helper()
	if bill.BillReference != reference {
		t.Fatalf("billReference = %q", bill.BillReference)
	}
	if compactPaymentCode(bill.WbcCode) != compactPaymentCode(paymentCode) {
		t.Fatalf("wbcCode = %q", bill.WbcCode)
	}
	if !sameDecimalString(bill.Amount, updatedAmount) {
		t.Fatalf("amount = %q", bill.Amount)
	}
	if bill.CustomerPhone != customerPhone {
		t.Fatalf("customerPhone = %q", bill.CustomerPhone)
	}
	if bill.UpdateTimeStamp == "" {
		t.Fatal("updateTimeStamp is empty")
	}
}

func looksLikePaymentCode(code string) bool {
	compact := compactPaymentCode(code)
	if len(compact) != 9 {
		return false
	}
	for _, char := range compact {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func compactPaymentCode(code string) string {
	return strings.ReplaceAll(code, " ", "")
}

func sameDecimalString(left, right string) bool {
	leftFloat, leftErr := strconv.ParseFloat(left, 64)
	rightFloat, rightErr := strconv.ParseFloat(right, 64)
	return leftErr == nil && rightErr == nil && leftFloat == rightFloat
}
