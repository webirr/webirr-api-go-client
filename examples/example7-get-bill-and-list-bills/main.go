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

	billReference := os.Getenv("WEBIRR_BILL_REFERENCE")
	if billReference != "" {
		response, err := client.GetBillByReference(context.Background(), billReference)
		if err != nil {
			panic(err)
		}
		if response.Error == "" {
			fmt.Println(response.Res.BillReference, response.Res.WbcCode, response.Res.UpdateTimeStamp)
		} else {
			fmt.Println("Error:", response.Error, response.ErrorCode)
		}
	}

	paymentCode := os.Getenv("WEBIRR_PAYMENT_CODE")
	if paymentCode != "" {
		response, err := client.GetBillByPaymentCode(context.Background(), paymentCode)
		if err != nil {
			panic(err)
		}
		if response.Error == "" {
			fmt.Println(response.Res.BillReference, response.Res.WbcCode, response.Res.UpdateTimeStamp)
		} else {
			fmt.Println("Error:", response.Error, response.ErrorCode)
		}
	}

	lastTimeStamp := "20251231"
	bills, err := client.GetBills(context.Background(), -1, lastTimeStamp, 100)
	if err != nil {
		panic(err)
	}
	if bills.Error != "" {
		fmt.Println("Error:", bills.Error, bills.ErrorCode)
		return
	}

	for _, bill := range bills.Res {
		fmt.Println(bill.BillReference, bill.WbcCode, bill.PaymentStatus, bill.UpdateTimeStamp)
	}
}
