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

// Data to send to LE (ACME)
type acmeMessage struct {
	Payload         string `json:"payload"`
	ProtectedHeader string `json:"protected"`
	Signature       string `json:"signature"`
}

// ProtectedHeader piece of the LE payload
type protectedHeader struct {
	Algorithm  string     `json:"alg"`
	JsonWebKey jsonWebKey `json:"jwk,omitempty"`
	KeyId      string     `json:"kid,omitempty"`
	Url        string     `json:"url"`
	Nonce      string     `json:"nonce"`
}

// JWK for the LE payload
type jsonWebKey struct {
	KeyType        string `json:"kty"`
	PublicExponent string `json:"e,omitempty"`   // RSA
	Modulus        string `json:"n,omitempty"`   // RSA
	CurveName      string `json:"crv,omitempty"` // EC
	CurvePointX    string `json:"x,omitempty"`   // EC
	CurvePointY    string `json:"y,omitempty"`   // EC
}

// LE account data
type AcmeAccount struct {
	Status               string   `json:"status,omitempty"`
	Contact              []string `json:"contact,omitempty"`
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed,omitempty"`
	OnlyReturnExisting   bool     `json:"onlyReturnExisting,omitempty"`
}

// LE response to account data post/update
type AcmeAccountResponse struct {
	JsonWebKey jsonWebKey `json:"key"`
	Contact    []string   `json:"contact"`
	Orders     string     `json:"orders"`
	InitialIP  string     `json:"initialIp"`
	CreatedAt  string     `json:"createdAt"`
	Status     string     `json:"status"`
	Location   string     `json:"-"` // omit because it is in the header
}
