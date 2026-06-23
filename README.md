# Official Go Client Library for WeBirr Payment Gateway APIs

This is the official Go client library for integrating merchant applications
with WeBirr Payment Gateway APIs.

## Installation

```bash
go get github.com/webirr/webirr-api-go-client
```

## Usage

For TestEnv examples, set these environment variables:

```bash
export WEBIRR_TEST_ENV_MERCHANT_ID=0305
export WEBIRR_TEST_ENV_API_KEY=your-test-env-api-key
```

Create a client for TestEnv:

```go
client := webirr.NewClient(
	os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
	os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
	true,
)
```

The merchant ID is configured on the client. When you create or update a bill,
the client sets `Bill.MerchantID` automatically when the configured merchant ID
is not empty.

You can also pass your own reusable `*http.Client`:

```go
httpClient := &http.Client{Timeout: 30 * time.Second}
client := webirr.NewClient(merchantID, apiKey, true, webirr.WithHTTPClient(httpClient))
```

## Creating/Updating bill

Create a bill:

```go
package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client"
)

func main() {
	client := webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)

	bill := &webirr.Bill{
		Amount:        "270.90",
		CustomerCode:  "cc01",
		CustomerName:  "Elias Haileselassie",
		CustomerPhone: "0911000000",
		Time:          "2021-07-22 22:14",
		Description:   "hotel booking",
		BillReference: "go/2021/132",
	}

	response, err := client.CreateBill(context.Background(), bill)
	if err != nil {
		panic(err)
	}

	if response.Error == "" {
		fmt.Println("WeBirr Payment Code:", response.Res)
	} else {
		fmt.Println("Error:", response.Error, response.ErrorCode)
	}
}
```

Update the same bill while it is unpaid:

```go
bill.Amount = "278.00"
bill.CustomerName = "Elias Haileselassie"

response, err := client.UpdateBill(context.Background(), bill)
if err != nil {
	panic(err)
}

if response.Error == "" {
	fmt.Println("Update result:", response.Res)
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

## Getting Bill and Listing Bills

Get a bill by merchant reference:

```go
response, err := client.GetBillByReference(context.Background(), "go/2021/132")
if err != nil {
	panic(err)
}

if response.Error == "" {
	fmt.Println(response.Res.WbcCode)
	fmt.Println(response.Res.UpdateTimeStamp)
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

Get a bill by WeBirr payment code:

```go
response, err := client.GetBillByPaymentCode(context.Background(), "123 456 789")
if err != nil {
	panic(err)
}
```

List bills by payment status and timestamp cursor:

```go
lastTimeStamp := "20251231" // Empty string starts from the beginning. Time parts can also be used.
response, err := client.GetBills(context.Background(), -1, lastTimeStamp, 100)
if err != nil {
	panic(err)
}

if response.Error == "" {
	for _, bill := range response.Res {
		fmt.Println(bill.BillReference, bill.WbcCode, bill.PaymentStatus, bill.UpdateTimeStamp)
	}
}
```

## Getting Supported Banks for Checkout

Use this endpoint to display only the banks and wallets configured for the
merchant.

```go
response, err := client.GetSupportedBanks(context.Background())
if err != nil {
	panic(err)
}

if response.Error == "" {
	for _, bank := range response.Res {
		fmt.Println(bank.BankID, bank.Name)
	}
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

## Getting Payment status

```go
response, err := client.GetPaymentStatus(context.Background(), "123 456 789")
if err != nil {
	panic(err)
}

if response.Error == "" {
	if response.Res.IsPaid() && response.Res.Data != nil {
		fmt.Println("Paid:", response.Res.Data.PaymentReference)
		fmt.Println("Paid at:", response.Res.Data.PaymentDate)
	} else {
		fmt.Println("Not paid yet")
	}
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

## sample object

```go
bill := &webirr.Bill{
	Amount:        "270.90",
	CustomerCode:  "cc01",
	CustomerName:  "Elias Haileselassie",
	CustomerPhone: "0911000000",
	Time:          "2021-07-22 22:14",
	Description:   "hotel booking",
	BillReference: "go/2021/132",
}
```

## Deleting bill

```go
response, err := client.DeleteBill(context.Background(), "123 456 789")
if err != nil {
	panic(err)
}

if response.Error == "" {
	fmt.Println("Delete result:", response.Res)
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

## Payment status bulk polling

Bulk polling returns payments updated after the supplied timestamp cursor.

```go
lastTimeStamp := "20251231" // Empty string starts from the beginning. Time parts can also be used.
response, err := client.GetPayments(context.Background(), lastTimeStamp, 100)
if err != nil {
	panic(err)
}

if response.Error == "" {
	for _, payment := range response.Res {
		fmt.Println(payment.WbcCode, payment.PaymentReference, payment.PaymentDate, payment.UpdateTimeStamp)
	}
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
}
```

## Webhooks - Payment processing using Webhook Callbacks

Merchants can receive payment notifications on their own webhook endpoint. The
payload can be decoded into `webirr.PaymentResponse`.

```go
func paymentWebhook(w http.ResponseWriter, r *http.Request) {
	var payment webirr.PaymentResponse
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if payment.IsPaid() {
		fmt.Println("Payment Reference:", payment.PaymentReference)
		fmt.Println("Paid at:", payment.PaymentDate)
	}

	w.WriteHeader(http.StatusOK)
}
```

## Gettting basic Statistics

```go
response, err := client.GetStat(context.Background(), "2021-01-01", "2021-12-31")
if err != nil {
	panic(err)
}

if response.Error == "" {
	fmt.Println("Bills:", response.Res.NBills)
	fmt.Println("Paid:", response.Res.NBillsPaid)
	fmt.Println("Amount Paid:", response.Res.AmountPaid)
} else {
	fmt.Println("Error:", response.Error, response.ErrorCode)
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
cd examples/example1-create-update-bill
go run .
```
