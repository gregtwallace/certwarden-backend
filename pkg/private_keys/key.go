package private_keys

import (
	"legocerthub-backend/pkg/utils"
)

// a single private key
type Key struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Algorithm   utils.Algorithm `json:"algorithm"`
	Pem         string          `json:"pem,omitempty"`
	ApiKey      string          `json:"api_key,omitempty"`
	CreatedAt   int             `json:"created_at,omitempty"`
	UpdatedAt   int             `json:"updated_at,omitempty"`
}

// key payload from PUT/POST
type KeyPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AlgorithmValue string `json:"algorithm_value"`
	PemContent     string `json:"pem"`
}

// new private key options
// used to return info about valid options when making a new key
type newKeyOptions struct {
	KeyAlgorithms []utils.Algorithm `json:"key_algorithms"`
}
