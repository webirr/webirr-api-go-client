package main

import (
	"context"
	"fmt"
	"os"

	webirr "github.com/webirr/webirr-api-go-client-"
)

func main() {
	client := webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)

	lastTimeStamp := "20251231"
	response, err := client.GetPayments(context.Background(), lastTimeStamp, 100)
	if err != nil {
		panic(err)
	}
	if response.Error != "" {
		fmt.Println("Error:", response.Error, response.ErrorCode)
		return
	}

	for _, payment := range response.Res {
		fmt.Println(payment.WbcCode, payment.PaymentReference, payment.PaymentDate, payment.UpdateTimeStamp)
	}
}
