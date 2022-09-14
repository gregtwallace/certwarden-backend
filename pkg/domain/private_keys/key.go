package private_keys

import (
	"crypto"
	"errors"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

var errBadKey = errors.New("bad crypto key")

// a single private key
type Key struct {
	ID           *int                  `json:"id,omitempty"`
	Name         *string               `json:"name,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Algorithm    *key_crypto.Algorithm `json:"algorithm,omitempty"`
	Pem          *string               `json:"pem,omitempty"`
	ApiKey       *string               `json:"api_key,omitempty"`
	ApiKeyViaUrl bool                  `json:"api_key_via_url,omitempty"`
	CreatedAt    *int                  `json:"created_at,omitempty"`
	UpdatedAt    *int                  `json:"updated_at,omitempty"`
}

// CryptoKey() returns the crypto.PrivateKey for a given key object
func (key *Key) CryptoKey() (cryptoKey crypto.PrivateKey, err error) {
	// nil pointer check
	if key == nil || key.Algorithm == nil || key.Pem == nil {
		return nil, errBadKey
	}

	// generate key from pem
	cryptoKey, err = key_crypto.PemStringToKey(*key.Pem, key.Algorithm.Value)
	if err != nil {
		return nil, err
	}

	return cryptoKey, nil
}

// new private key options
// used to return info about valid options when making a new key
type newKeyOptions struct {
	KeyAlgorithms []key_crypto.Algorithm `json:"key_algorithms"`
}
