package private_keys

import (
	"database/sql"
	"log"
	"time"
)

type KeyAppDb struct {
	Database *sql.DB
	Timeout  time.Duration
}

// PrivateKeys struct for database access
type KeysApp struct {
	Logger *log.Logger
	DB     KeyAppDb
}

// a single private key
type Key struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Algorithm   algorithm `json:"algorithm"`
	Pem         string    `json:"pem,omitempty"`
	ApiKey      string    `json:"api_key,omitempty"`
	CreatedAt   int       `json:"created_at,omitempty"`
	UpdatedAt   int       `json:"updated_at,omitempty"`
}

// a single private key, as database table fields
type KeyDb struct {
	ID             int
	Name           string
	Description    sql.NullString
	AlgorithmValue string
	Pem            string
	ApiKey         string
	CreatedAt      int
	UpdatedAt      int
}

// translate the db object into the api object
func (keyDb *KeyDb) KeyDbToKey() Key {
	return Key{
		ID:          keyDb.ID,
		Name:        keyDb.Name,
		Description: keyDb.Description.String,
		Algorithm:   algorithmByValue(keyDb.AlgorithmValue),
		Pem:         keyDb.Pem,
		ApiKey:      keyDb.ApiKey,
		CreatedAt:   keyDb.CreatedAt,
		UpdatedAt:   keyDb.UpdatedAt,
	}
}

// private key payload from PUT/POST
type privateKeyPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AlgorithmValue string `json:"algorithm_value"`
	PemContent     string `json:"pem"`
}

// new private key options
// used to return info about valid options when making a new key
type newPrivateKeyOptions struct {
	KeyAlgorithms []algorithm `json:"key_algorithms"`
}
