package acme_utils

// Directory struct that holds the returned data from querying directory URL
type AcmeDirectory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	NewAuthz   string `json:"newAuthz"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       struct {
		CaaIdentities  []string `json:"caaIdentities"`
		TermsOfService string   `json:"termsOfService"`
		Website        string   `json:"website"`
	} `json:"meta"`
}
