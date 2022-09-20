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
// ExtendedKey object. if includePem is false, the pem will be
// omitted from the KeyExtended object.
func (key keyDbExtended) toKeyExtended(includePem bool) private_keys.KeyExtended {
	keyExt := private_keys.KeyExtended{}

	// translate the regular Key fields
	keyExt.Key = key.toKey()

	// translate the extended fields
	// include sensitive info?
	if includePem {
		keyExt.Pem = key.pem
	}

	keyExt.ApiKey = key.apiKey
	keyExt.ApiKeyViaUrl = key.apiKeyViaUrl
	keyExt.CreatedAt = key.createdAt
	keyExt.UpdatedAt = key.updatedAt

	return keyExt
}
