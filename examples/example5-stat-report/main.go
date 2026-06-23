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

	response, err := client.GetStat(context.Background(), "2021-01-01", "2021-12-31")
	if err != nil {
		panic(err)
	}
	if response.Error != "" {
		fmt.Println("Error:", response.Error, response.ErrorCode)
		return
	}

	fmt.Println("Bills:", response.Res.NBills)
	fmt.Println("Paid:", response.Res.NBillsPaid)
	fmt.Println("Amount Paid:", response.Res.AmountPaid)
}
