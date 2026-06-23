package webirr

import "encoding/json"

// PaymentStatus is returned by the single payment-status endpoint.
type PaymentStatus struct {
	Status int            `json:"status"`
	Data   *PaymentDetail `json:"data"`
}

func (p PaymentStatus) IsPaid() bool {
	return p.Status == 2
}

// PaymentDetail contains paid bill details returned by single payment status.
type PaymentDetail struct {
	Status           int    `json:"status"`
	ID               int64  `json:"id"`
	BankID           string `json:"bankID"`
	PaymentReference string `json:"paymentReference"`
	PaymentDate      string `json:"paymentDate"`
	Time             string `json:"time"`
	Confirmed        bool   `json:"confirmed"`
	ConfirmedTime    string `json:"confirmedTime"`
	Canceled         bool   `json:"canceled"`
	CanceledTime     string `json:"canceledTime"`
	Amount           string `json:"amount"`
	WbcCode          string `json:"wbcCode"`
	UpdateTimeStamp  string `json:"updateTimeStamp"`
}

func (p PaymentDetail) IsPaid() bool {
	return p.Status == 2
}

func (p PaymentDetail) IsReversed() bool {
	return p.Status == 3 || p.Canceled
}

func (p *PaymentDetail) UnmarshalJSON(data []byte) error {
	type paymentDetailJSON PaymentDetail
	var out paymentDetailJSON
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	if out.PaymentDate == "" {
		out.PaymentDate = out.Time
	}
	if out.Time == "" {
		out.Time = out.PaymentDate
	}
	*p = PaymentDetail(out)
	return nil
}

// PaymentResponse is returned by bulk polling and webhook payloads.
type PaymentResponse struct {
	Status           int    `json:"status"`
	ID               int64  `json:"id"`
	BankID           string `json:"bankID"`
	PaymentReference string `json:"paymentReference"`
	PaymentDate      string `json:"paymentDate"`
	Time             string `json:"time"`
	Confirmed        bool   `json:"confirmed"`
	ConfirmedTime    string `json:"confirmedTime"`
	Canceled         bool   `json:"canceled"`
	CanceledTime     string `json:"canceledTime"`
	Amount           string `json:"amount"`
	WbcCode          string `json:"wbcCode"`
	UpdateTimeStamp  string `json:"updateTimeStamp"`
}

func (p PaymentResponse) IsPaid() bool {
	return p.Status == 2
}

func (p PaymentResponse) IsReversed() bool {
	return p.Status == 3 || p.Canceled
}

func (p *PaymentResponse) UnmarshalJSON(data []byte) error {
	type paymentResponseJSON PaymentResponse
	var out paymentResponseJSON
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	if out.PaymentDate == "" {
		out.PaymentDate = out.Time
	}
	if out.Time == "" {
		out.Time = out.PaymentDate
	}
	*p = PaymentResponse(out)
	return nil
}
