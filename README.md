Official Go Client Library for WeBirr Payment Gateway APIs

This Client Library provides convenient access to WeBirr Payment Gateway APIs from Go Applications.

## Install

```bash
go get github.com/webirr/webirr-api-go-client
```

## Usage

The library needs to be configured with a *merchant Id* & *API key*. You can get it by contacting [webirr.com](https://webirr.net)

> You can use this library for production or test environments. you will need to set isTestEnv=true for test, and false for production apps when creating a `webirr.Client`

Examples assume the WeBirr TestEnv and read credentials from environment variables:

```bash
export WEBIRR_TEST_ENV_MERCHANT_ID="YOUR_TEST_MERCHANT_ID"
export WEBIRR_TEST_ENV_API_KEY="YOUR_TEST_API_KEY"
```

Create the client with merchant ID, API key, and environment once. The client automatically sets `Bill.MerchantID` before sending bill create/update requests when the configured merchant ID is not empty, so application code and examples should not set `MerchantID` on the bill object.

For batch jobs, overnight bill uploads, and polling workers, pass your own reusable `*http.Client`:

```go
httpClient := &http.Client{Timeout: 30 * time.Second}
api := webirr.NewClient(merchantId, apiKey, true, webirr.WithHTTPClient(httpClient))
```

## Error handling & retries

WeBirr business errors come back on an HTTP 2xx response in `ApiResponse.Error` / `ApiResponse.ErrorCode`, such as invalid API key, duplicate bill reference, or validation errors. Everything else is a platform error returned through Go's normal `error` channel: network/DNS/TLS failures, timeouts, non-2xx HTTP, and empty or non-JSON 2xx bodies.

Retry only transient platform failures with exponential backoff and jitter: connection errors, timeouts, and HTTP 5xx / 429 / 408. Non-2xx responses return `*webirr.HTTPError`, and `webirr.IsTransient(err)` classifies HTTP status errors, `context.DeadlineExceeded`, and `net.Error` timeouts. Never retry other 4xx responses.

Create and read operations are safe to retry. `DeleteBill` is also safe to retry, but a retry after it already succeeded returns an "invalid payment code" business error; treat that as already deleted.

```go
response, err := api.CreateBill(context.Background(), bill)
if err != nil {
	if webirr.IsTransient(err) {
		// retry with backoff + jitter
	}
	// handle platform error
	return
}

if response.Error != "" {
	// WeBirr business error from a 2xx response envelope.
	fmt.Println("error:", response.Error)
	fmt.Println("errorCode:", response.ErrorCode)
	return
}

fmt.Println("Payment Code =", response.Res)
```

## Example

The examples below keep the usual WeBirr SDK flow: create the client, call the API, check `Error`, handle the success branch, and print `ErrorCode` on failure.

### Creating a new Bill / Updating an existing Bill on WeBirr Servers

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	webirr "github.com/webirr/webirr-api-go-client"
)

func main() {
	apiKey := os.Getenv("WEBIRR_TEST_ENV_API_KEY")
	merchantId := os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID")

	//apiKey := "YOUR_API_KEY"
	//merchantId := "YOUR_MERCHANT_ID"

	api := webirr.NewClient(merchantId, apiKey, true)

	bill := &webirr.Bill{
		Amount:        "270.90",
		CustomerCode:  "cc01", // it can be email address or phone number if you dont have customer code
		CustomerName:  "Elias Haileselassie",
		CustomerPhone: "0911000000", // optional; used for SMS notification when enabled for the merchant
		Time:          "2021-07-22 22:14", // your bill time, always in this format
		Description:   "hotel booking",
		BillReference: "go/example/" + time.Now().UTC().Format("20060102150405"), // your unique reference number
	}

	fmt.Println("Creating Bill...")

	res, err := api.CreateBill(context.Background(), bill)
	if err != nil {
		panic(err)
	}

	if res.Error == "" {
		// success
		paymentCode := res.Res // returns paymentcode such as 429 723 975
		fmt.Println("Payment Code =", paymentCode) // we may want to save payment code in local db.
	} else {
		// fail
		fmt.Println("error:", res.Error)
		fmt.Println("errorCode:", res.ErrorCode) // can be used to handle specific business error such as ERROR_INVLAID_INPUT_DUP_REF
	}

	// Update existing bill if it is not paid
	bill.Amount = "278.00"
	bill.CustomerName = "Elias go"
	//bill.BillReference = "WE CAN NOT CHANGE THIS"

	fmt.Println("Updating Bill...")

	res, err = api.UpdateBill(context.Background(), bill)
	if err != nil {
		panic(err)
	}

	if res.Error == "" {
		// success
		fmt.Println("bill is updated successfully") // res.Res will be "OK"; no need to check here.
	} else {
		// fail
		fmt.Println("error:", res.Error)
		fmt.Println("errorCode:", res.ErrorCode) // can be used to handle specific business error such as ERROR_INVLAID_INPUT
	}
}
```

### Getting a Bill and Listing Bills

```go
package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client"
)

func main() {
	api := webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)

	billReference := "YOUR_BILL_REFERENCE" // BILL_REFERENCE_YOU_SAVED_AFTER_CREATING_A_NEW_BILL
	paymentCode := "YOUR_PAYMENT_CODE"     // PAYMENT_CODE_YOU_SAVED_AFTER_CREATING_A_NEW_BILL

	fmt.Println("Getting bill by reference...")
	resByReference, err := api.GetBillByReference(context.Background(), billReference)
	if err != nil {
		panic(err)
	}
	if resByReference.Error == "" {
		// success
		fmt.Println("Bill Reference:", resByReference.Res.BillReference)
		fmt.Println("Payment Code:", resByReference.Res.WbcCode)
		fmt.Println("Amount:", resByReference.Res.Amount)
		fmt.Println("Payment Status:", resByReference.Res.PaymentStatus)
		fmt.Println("Update Timestamp:", resByReference.Res.UpdateTimeStamp)
	} else {
		// fail
		fmt.Println("error:", resByReference.Error)
		fmt.Println("errorCode:", resByReference.ErrorCode)
	}

	fmt.Println("Getting bill by payment code...")
	resByCode, err := api.GetBillByPaymentCode(context.Background(), paymentCode)
	if err != nil {
		panic(err)
	}
	if resByCode.Error == "" {
		// success
		fmt.Println("Bill Reference:", resByCode.Res.BillReference)
		fmt.Println("Payment Code:", resByCode.Res.WbcCode)
	} else {
		// fail
		fmt.Println("error:", resByCode.Error)
		fmt.Println("errorCode:", resByCode.ErrorCode)
	}

	fmt.Println("Listing bills...")
	paymentStatus := -1       // -1 all, 0 pending, 1 unconfirmed payment, 2 paid.
	lastTimeStamp := "20251231" // Date-only cursor; use "20251231235959" when you need time precision.
	limit := 100

	bills, err := api.GetBills(context.Background(), paymentStatus, lastTimeStamp, limit)
	if err != nil {
		panic(err)
	}
	if bills.Error == "" {
		// success
		fmt.Println("Bills returned:", len(bills.Res))
		for _, bill := range bills.Res {
			fmt.Println("-----------------------------")
			fmt.Println("Bill Reference:", bill.BillReference)
			fmt.Println("Payment Code:", bill.WbcCode)
			fmt.Println("Amount:", bill.Amount)
			fmt.Println("Payment Status:", bill.PaymentStatus)
			fmt.Println("Update Timestamp:", bill.UpdateTimeStamp)
		}
	} else {
		// fail
		fmt.Println("error:", bills.Error)
		fmt.Println("errorCode:", bills.ErrorCode)
	}
}
```

Timestamp cursors can be date-only (`yyyyMMdd`) or include time (`yyyyMMddHHmmss`). Use empty string only when you intentionally want all history from the beginning.

### Getting Supported Banks for Checkout

Use this endpoint to display only the banks and wallets configured for the merchant.

```go
response, err := api.GetSupportedBanks(context.Background())
if err != nil {
	panic(err)
}

if response.Error == "" {
	for _, bank := range response.Res {
		fmt.Println(bank.BankID, "-", bank.Name)
	}
	fmt.Println("Use only these merchant-specific banks when showing checkout payment instructions.")
} else {
	fmt.Println("error:", response.Error)
	fmt.Println("errorCode:", response.ErrorCode)
}
```

Checkout pages should render bank-specific instructions only from `GetSupportedBanks()`. Do not show a broad static bank list unless those banks are returned for the configured merchant.

### Getting Payment status of an existing Bill from WeBirr Servers

```go
package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client"
)

func main() {
	api := webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)

	paymentCode := "PAYMENT_CODE_YOU_SAVED_AFTER_CREATING_A_NEW_BILL" // such as "141 263 782"

	fmt.Println("Getting Payment Status...")

	res, err := api.GetPaymentStatus(context.Background(), paymentCode)
	if err != nil {
		panic(err)
	}

	if res.Error == "" {
		// success
		if res.Res.IsPaid() && res.Res.Data != nil { // 0. Pending, 1. Payment in Progress, 2. Paid
			payment := res.Res.Data
			fmt.Println("bill is paid")
			fmt.Println("bill payment detail")
			fmt.Println("Bank:", payment.BankID)
			fmt.Println("Bank Reference Number:", payment.PaymentReference)
			fmt.Println("Amount Paid:", payment.Amount)
			fmt.Println("Payment Date:", payment.PaymentDate)
		} else {
			fmt.Println("bill is pending payment")
		}
	} else {
		// fail
		fmt.Println("error:", res.Error)
		fmt.Println("errorCode:", res.ErrorCode) // can be used to handle specific business error such as ERROR_INVLAID_INPUT
	}
}
```

*Sample object returned from GetPaymentStatus()*

```javascript
{
  error: null,
  res: {
    status: 2,
    data: {
      status: 2,
      id: 111219507,
      bankID: "cbe_mobile",
      paymentReference: "TX70e78862148f4c249606",
      paymentDate: "2025-02-26 22:17:19",
      confirmed: true,
      confirmedTime: "2025-02-26 22:17:19",
      amount: "278",
      wbcCode: "149 233 514",
      updateTimeStamp: "2025022622171981338"
    }
  },
  errorCode: null
}
```

### Deleting an existing Bill from WeBirr Servers (if it is not paid)

```go
response, err := api.DeleteBill(context.Background(), "PAYMENT_CODE_YOU_SAVED_AFTER_CREATING_A_NEW_BILL")
if err != nil {
	panic(err)
}

if response.Error == "" {
	// success
	fmt.Println("bill is deleted successfully")
} else {
	// fail
	fmt.Println("error:", response.Error)
	fmt.Println("errorCode:", response.ErrorCode)
}
```

### Payment status bulk polling

Use timestamp-based polling to synchronize paid or reversed payments in batch jobs.

```go
lastTimeStamp := "20251231" // Date-only cursor; use "20251231235959" when you need time precision.
limit := 100

response, err := api.GetPayments(context.Background(), lastTimeStamp, limit)
if err != nil {
	panic(err)
}

if response.Error == "" {
	nextLastTimeStamp := lastTimeStamp
	for _, payment := range response.Res {
		fmt.Println("-----------------------------")
		fmt.Println("Payment Code:", payment.WbcCode)
		fmt.Println("Bank Reference Number:", payment.PaymentReference)
		fmt.Println("Amount Paid:", payment.Amount)
		fmt.Println("Payment Date:", payment.PaymentDate)
		fmt.Println("Update Timestamp:", payment.UpdateTimeStamp)

		if payment.UpdateTimeStamp > nextLastTimeStamp {
			nextLastTimeStamp = payment.UpdateTimeStamp
		}
	}

	// Persist nextLastTimeStamp only after the batch is processed successfully.
	fmt.Println("Next cursor:", nextLastTimeStamp)
} else {
	fmt.Println("error:", response.Error)
	fmt.Println("errorCode:", response.ErrorCode)
}
```

Do not use obsolete serial-number polling for new integrations.

### Webhooks - Payment processing using Webhook Callbacks

Merchants can receive payment notifications on their own HTTPS webhook endpoint. The endpoint should validate the request method, check the configured `authKey` when used, decode the raw JSON payload into `webirr.PaymentResponse`, process the payment idempotently, and return success quickly.

```go
func paymentWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expectedAuthKey := os.Getenv("WEBIRR_WEBHOOK_AUTH_KEY")
	if expectedAuthKey != "" && r.URL.Query().Get("authKey") != expectedAuthKey {
		http.Error(w, "invalid authKey", http.StatusUnauthorized)
		return
	}

	var payment webirr.PaymentResponse
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if payment.IsPaid() {
		fmt.Println("Payment Reference:", payment.PaymentReference)
		fmt.Println("Paid Via:", payment.BankID)
		fmt.Println("Amount Paid:", payment.Amount)
		fmt.Println("Payment Date:", payment.PaymentDate)
	} else if payment.IsReversed() {
		fmt.Println("Payment reversed:", payment.PaymentReference)
	} else {
		fmt.Println("Payment status:", payment.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

Webhook processing and polling should share the same local completion logic so that repeated callbacks, manual polling, or background reconciliation cannot complete the same merchant order twice.

### Getting basic Statistics

```go
response, err := api.GetStat(context.Background(), "2021-01-01", "2021-12-31")
if err != nil {
	panic(err)
}

if response.Error == "" {
	fmt.Println("Bills:", response.Res.NBills)
	fmt.Println("Paid:", response.Res.NBillsPaid)
	fmt.Println("Amount Paid:", response.Res.AmountPaid)
} else {
	fmt.Println("error:", response.Error)
	fmt.Println("errorCode:", response.ErrorCode)
}
```

## Examples

The `examples` directory contains runnable examples for:

- Creating and updating bills
- Single payment-status polling
- Deleting a bill
- Timestamp-based bulk payment polling
- Merchant statistics
- Webhook callback handling
- Getting and listing bills
- Getting merchant-supported banks

Run an example:

```bash
go run ./examples/example1-create-update-bill
```

Run tests:

```bash
go test ./...
```

Live TestEnv smoke tests run only when these environment variables are set:

```bash
export WEBIRR_TEST_ENV_MERCHANT_ID="YOUR_TEST_MERCHANT_ID"
export WEBIRR_TEST_ENV_API_KEY="YOUR_TEST_API_KEY"
go test ./...
```
