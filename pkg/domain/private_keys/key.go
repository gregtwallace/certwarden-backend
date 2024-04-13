package private_keys

import (
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"crypto"
	"fmt"
	"time"
)

// Key is a single private key with all data
type Key struct {
	ID             int
	Name           string
	Description    string
	Algorithm      key_crypto.Algorithm
	Pem            string
	ApiKey         string
	ApiKeyNew      string
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
	ApiKey    string `json:"api_key"`
	ApiKeyNew string `json:"api_key_new,omitempty"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
	// exclude PEM
}

func (key Key) detailedResponse() keyDetailedResponse {
	return keyDetailedResponse{
		KeySummaryResponse: key.SummaryResponse(),

		ApiKey:    key.ApiKey,
		ApiKeyNew: key.ApiKeyNew,
		CreatedAt: key.CreatedAt,
		UpdatedAt: key.UpdatedAt,
	}
}

// Output Methods

func (key Key) FilenameNoExt() string {
	return fmt.Sprintf("%s.key.pem", key.Name)
}

func (key Key) PemContent() string {
	return key.Pem
}

func (key Key) Modtime() time.Time {
	return time.Unix(int64(key.UpdatedAt), 0)
}

// end Output Methods

// CryptoPrivateKey() provides a crypto.PrivateKey for the Key
// for the Account
func (key *Key) CryptoPrivateKey() (cryptoKey crypto.PrivateKey, err error) {
	return (key_crypto.PemStringToKey(key.Pem, key.Algorithm))
}
