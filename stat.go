package webirr

import "encoding/json"

// Stat contains merchant bill/payment statistics for a date range.
type Stat struct {
	NBills       int    `json:"nBills"`
	NBillsPaid   int    `json:"nBillsPaid"`
	NBillsUnpaid int    `json:"nBillsUnpaid"`
	AmountBills  string `json:"amountBills"`
	AmountPaid   string `json:"amountPaid"`
	AmountUnpaid string `json:"amountUnpaid"`
}

func (s *Stat) UnmarshalJSON(data []byte) error {
	type statJSON struct {
		NBills         int    `json:"nBills"`
		NBillsUpper    int    `json:"NBills"`
		NBillsPaid     int    `json:"nBillsPaid"`
		NBillsPaidUp   int    `json:"NBillsPaid"`
		NBillsUnpaid   int    `json:"nBillsUnpaid"`
		NBillsUnpaidUp int    `json:"NBillsUnpaid"`
		AmountBills    string `json:"amountBills"`
		AmountBillsUp  string `json:"AmountBills"`
		AmountPaid     string `json:"amountPaid"`
		AmountPaidUp   string `json:"AmountPaid"`
		AmountUnpaid   string `json:"amountUnpaid"`
		AmountUnpaidUp string `json:"AmountUnpaid"`
	}
	var out statJSON
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	s.NBills = firstInt(out.NBills, out.NBillsUpper)
	s.NBillsPaid = firstInt(out.NBillsPaid, out.NBillsPaidUp)
	s.NBillsUnpaid = firstInt(out.NBillsUnpaid, out.NBillsUnpaidUp)
	s.AmountBills = firstString(out.AmountBills, out.AmountBillsUp)
	s.AmountPaid = firstString(out.AmountPaid, out.AmountPaidUp)
	s.AmountUnpaid = firstString(out.AmountUnpaid, out.AmountUnpaidUp)
	return nil
}

func firstInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func firstString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
