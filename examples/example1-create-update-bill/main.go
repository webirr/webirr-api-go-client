package main

import (
	"context"
	"fmt"
	"os"
	"time"

	webirr "github.com/webirr/webirr-api-go-client-"
)

func main() {
	client := testEnvClient()
	billReference := fmt.Sprintf("go/example/%s", time.Now().UTC().Format("20060102150405"))
	bill := &webirr.Bill{
		Amount:        "270.90",
		CustomerCode:  "cc01",
		CustomerName:  "Elias Haileselassie",
		CustomerPhone: "0911000000",
		Time:          "2021-07-22 22:14",
		Description:   "hotel booking",
		BillReference: billReference,
	}

	create, err := client.CreateBill(context.Background(), bill)
	if err != nil {
		panic(err)
	}
	if create.Error != "" {
		fmt.Println("Create error:", create.Error, create.ErrorCode)
		return
	}

	fmt.Println("WeBirr Payment Code:", create.Res)

	bill.Amount = "278.00"
	update, err := client.UpdateBill(context.Background(), bill)
	if err != nil {
		panic(err)
	}
	if update.Error == "" {
		fmt.Println("Update result:", update.Res)
	} else {
		fmt.Println("Update error:", update.Error, update.ErrorCode)
	}
}

func testEnvClient() *webirr.Client {
	return webirr.NewClient(
		os.Getenv("WEBIRR_TEST_ENV_MERCHANT_ID"),
		os.Getenv("WEBIRR_TEST_ENV_API_KEY"),
		true,
	)
}
