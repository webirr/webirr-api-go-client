package webirr

import "encoding/json"

// Bill represents the create/update request model for a WeBirr bill.
type Bill struct {
	Amount        string            `json:"amount"`
	CustomerCode  string            `json:"customerCode"`
	CustomerName  string            `json:"customerName"`
	CustomerPhone string            `json:"customerPhone"`
	Time          string            `json:"time"`
	Description   string            `json:"description"`
	BillReference string            `json:"billReference"`
	MerchantID    string            `json:"merchantID"`
	Extras        map[string]string `json:"extras"`
}

// MarshalJSON keeps empty extras as an object and customerPhone as a string.
func (b Bill) MarshalJSON() ([]byte, error) {
	type billJSON Bill
	out := billJSON(b)
	if out.Extras == nil {
		out.Extras = map[string]string{}
	}
	return json.Marshal(out)
}

// BillResponse represents a bill returned from lookup/list APIs.
type BillResponse struct {
	Amount          string            `json:"amount"`
	CustomerCode    string            `json:"customerCode"`
	CustomerName    string            `json:"customerName"`
	CustomerPhone   string            `json:"customerPhone"`
	Time            string            `json:"time"`
	Description     string            `json:"description"`
	BillReference   string            `json:"billReference"`
	MerchantID      string            `json:"merchantID"`
	Extras          map[string]string `json:"extras"`
	WbcCode         string            `json:"wbcCode"`
	PaymentStatus   int               `json:"paymentStatus"`
	UpdateTimeStamp string            `json:"updateTimeStamp"`
}
