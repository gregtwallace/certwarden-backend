package sqlite

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"time"
)

// keyDb is a single private key, as database table fields
// corresponds to private_keys.Key
type keyDb struct {
	id             int
	name           string
	description    string
	algorithmValue string
	pem            string
	apiKey         string
	apiKeyNew      string
	apiKeyDisabled bool
	apiKeyViaUrl   bool
	lastAccess     int64
	createdAt      int64
	updatedAt      int64
}

// toKey maps the database key info to the private_keys Key
// object
func (key keyDb) toKey() private_keys.Key {
	return private_keys.Key{
		ID:             key.id,
		Name:           key.name,
		Description:    key.description,
		Algorithm:      key_crypto.AlgorithmByStorageValue(key.algorithmValue),
		Pem:            key.pem,
		ApiKey:         key.apiKey,
		ApiKeyNew:      key.apiKeyNew,
		ApiKeyDisabled: key.apiKeyDisabled,
		ApiKeyViaUrl:   key.apiKeyViaUrl,
		LastAccess:     time.Unix(key.lastAccess, 0),
		CreatedAt:      time.Unix(key.createdAt, 0),
		UpdatedAt:      time.Unix(key.updatedAt, 0),
	}
}
