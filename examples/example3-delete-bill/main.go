package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client"
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

	response, err := client.DeleteBill(context.Background(), paymentCode)
	if err != nil {
		panic(err)
	}
	if response.Error == "" {
		fmt.Println("Delete result:", response.Res)
	} else {
		fmt.Println("Error:", response.Error, response.ErrorCode)
	}
}
