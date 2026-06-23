package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	webirr "github.com/webirr/webirr-api-go-client"
)

func main() {
	http.HandleFunc("/webirr/payment", paymentWebhook)
	fmt.Println("Listening on http://localhost:8080/webirr/payment")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func paymentWebhook(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println("Paid at:", payment.PaymentDate)
	} else {
		fmt.Println("Payment status:", payment.Status)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
