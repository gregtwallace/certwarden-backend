package private_keys

import (
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

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

// new private key options
// used to return info about valid options when making a new key
type newKeyOptions struct {
	KeyAlgorithms []key_crypto.Algorithm `json:"key_algorithms"`
}
