package sqlite

import (
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
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
	apiKeyDisabled bool
	apiKeyViaUrl   bool
	createdAt      int
	updatedAt      int
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
		ApiKeyDisabled: key.apiKeyDisabled,
		ApiKeyViaUrl:   key.apiKeyViaUrl,
		CreatedAt:      key.createdAt,
		UpdatedAt:      key.updatedAt,
	}
}
