package private_keys

import (
	"crypto"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// Key is a single private key with all data
type Key struct {
	ID             int
	Name           string
	Description    string
	Algorithm      key_crypto.Algorithm
	Pem            string
	ApiKey         string
	ApiKeyDisabled bool
	ApiKeyViaUrl   bool
	CreatedAt      int
	UpdatedAt      int
}

// keySummaryResponse is a JSON response containing only
// fields desired for the summary
type KeySummaryResponse struct {
	ID             int                  `json:"id"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Algorithm      key_crypto.Algorithm `json:"algorithm"`
	ApiKeyDisabled bool                 `json:"api_key_disabled"`
	ApiKeyViaUrl   bool                 `json:"api_key_via_url"`
}

func (key Key) SummaryResponse() KeySummaryResponse {
	return KeySummaryResponse{
		ID:             key.ID,
		Name:           key.Name,
		Description:    key.Description,
		Algorithm:      key.Algorithm,
		ApiKeyDisabled: key.ApiKeyDisabled,
		ApiKeyViaUrl:   key.ApiKeyViaUrl,
	}
}

// keyDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type keyDetailedResponse struct {
	KeySummaryResponse
	ApiKey    string `json:"api_key,omitempty"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
	// exclude PEM
}

func (key Key) detailedResponse(withSensitive bool) keyDetailedResponse {
	// option to redact sensitive info
	apiKey := key.ApiKey
	if !withSensitive {
		apiKey = "[redacted]"
	}

	return keyDetailedResponse{
		KeySummaryResponse: key.SummaryResponse(),

		ApiKey:    apiKey,
		CreatedAt: key.CreatedAt,
		UpdatedAt: key.UpdatedAt,
	}
}

// new private key options
// used to return info about valid options when making a new key
type newKeyOptions struct {
	KeyAlgorithms []key_crypto.Algorithm `json:"key_algorithms"`
}

// CryptoPrivateKey() provides a crypto.PrivateKey for the Key
// for the Account
func (key *Key) CryptoPrivateKey() (cryptoKey crypto.PrivateKey, err error) {
	return (key_crypto.PemStringToKey(key.Pem, key.Algorithm))
}
