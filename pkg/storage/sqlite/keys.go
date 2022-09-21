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
}

// toKey maps the database key info to the private_keys Key
// object
func (key keyDb) toKey() private_keys.Key {
	return private_keys.Key{
		ID:          key.id,
		Name:        key.name,
		Description: key.description,
		Algorithm:   key_crypto.AlgorithmByValue(key.algorithmValue),
	}
}

// keyDbExtended is a single private key, as database table
// fields with all fields represented
// corresponds to private_keys.KeyExtended
type keyDbExtended struct {
	keyDb
	pem          string
	apiKey       string
	apiKeyViaUrl bool
	createdAt    int
	updatedAt    int
}

// toKeyExtended maps the database key info to the private_keys
// ExtendedKey object.
func (key keyDbExtended) toKeyExtended() private_keys.KeyExtended {
	return private_keys.KeyExtended{
		// regular Key fields
		Key: key.toKey(),
		// Extended fields
		Pem:          key.pem,
		ApiKey:       key.apiKey,
		ApiKeyViaUrl: key.apiKeyViaUrl,
		CreatedAt:    key.createdAt,
		UpdatedAt:    key.updatedAt,
	}
}
