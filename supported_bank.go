package webirr

// SupportedBank is a bank or wallet configured for the merchant.
type SupportedBank struct {
	BankID string `json:"bankID"`
	Name   string `json:"name"`
}
