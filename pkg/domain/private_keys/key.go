package private_keys

import (
	"crypto"
	"fmt"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
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

// Pem Output Methods

// PemFilename returns the filename that should be sent to the client when Key
// is sent to the client in Pem format
func (key Key) PemFilename() string {
	return fmt.Sprintf("%s.key.pem", key.Name)
}

// PemContent returns the actual Pem data of the private key
func (key Key) PemContent() string {
	return key.Pem
}

// PemModtime returns the time the key resource was last updated at. Note: it is
// possible for modtime to become newer without the pem actually changing
// if other information on the key object was modified. This is acceptable
// and preferred over other methods I considered to avoid this.
// For example, since the Key PEM can never be modified, could use the created at
// time instead. However, if cert's key is changed but the name is renamed back to
// the name of the first key, the pem has effectively changed and if the 'new' key
// actually has an older created timestamp this Modtime would be wrong and not signal
// the need for an update.
func (key Key) PemModtime() time.Time {
	return time.Unix(int64(key.UpdatedAt), 0)
}

// end Pem Output Methods

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
