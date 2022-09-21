package private_keys

import (
	"crypto"
	"errors"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

var errBadKey = errors.New("bad crypto key")

// Key is a single private key (summary for all keys)
type Key struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Algorithm   key_crypto.Algorithm `json:"algorithm"`
}

// KeyExtended is a single private key with all of its details
type KeyExtended struct {
	Key
	Pem          string `json:"-"`
	ApiKey       string `json:"api_key,omitempty"`
	ApiKeyViaUrl bool   `json:"api_key_via_url"`
	CreatedAt    int    `json:"created_at"`
	UpdatedAt    int    `json:"updated_at"`
}

// CryptoKey() returns the crypto.PrivateKey for a given key object
func (key *KeyExtended) CryptoKey() (cryptoKey crypto.PrivateKey, err error) {
	// nil pointer check
	if key == nil || key.Algorithm == key_crypto.UnknownAlgorithm || key.Pem == "" {
		return nil, errBadKey
	}

	// generate key from pem
	cryptoKey, err = key_crypto.PemStringToKey(key.Pem, key.Algorithm)
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
