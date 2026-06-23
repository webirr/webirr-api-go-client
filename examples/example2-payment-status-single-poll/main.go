package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client-"
)

func main() {
	paymentCode := os.Getenv("WEBIRR_PAYMENT_CODE")
	if paymentCode == "" {
		fmt.Println("Set WEBIRR_PAYMENT_CODE to a WeBirr payment code.")
		return
	}

	client := webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)

	response, err := client.GetPaymentStatus(context.Background(), paymentCode)
	if err != nil {
		panic(err)
	}
	if response.Error != "" {
		fmt.Println("Error:", response.Error, response.ErrorCode)
		return
	}

	if response.Res.IsPaid() && response.Res.Data != nil {
		fmt.Println("Paid:", response.Res.Data.PaymentReference)
		fmt.Println("Paid at:", response.Res.Data.PaymentDate)
		return
	}

	fmt.Println("Not paid yet")
}
