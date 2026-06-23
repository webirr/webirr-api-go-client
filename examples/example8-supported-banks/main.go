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

	response, err := client.GetSupportedBanks(context.Background())
	if err != nil {
		panic(err)
	}
	if response.Error != "" {
		fmt.Println("Error:", response.Error, response.ErrorCode)
		return
	}

	for _, bank := range response.Res {
		fmt.Println(bank.BankID, bank.Name)
	}
}
