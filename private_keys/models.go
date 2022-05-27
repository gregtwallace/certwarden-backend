package private_keys

import (
	"database/sql"
	"log"
	"time"
)

// PrivateKeys struct for database access
type PrivateKeysApp struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
}

// a single private key
type PrivateKey struct {
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
type PrivateKeyDb struct {
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
func (privateKeyDb *PrivateKeyDb) PrivateKeyDbToPk() *PrivateKey {
	return &PrivateKey{
		ID:          privateKeyDb.ID,
		Name:        privateKeyDb.Name,
		Description: privateKeyDb.Description.String,
		Algorithm:   algorithmByValue(privateKeyDb.AlgorithmValue),
		Pem:         privateKeyDb.Pem,
		ApiKey:      privateKeyDb.ApiKey,
		CreatedAt:   privateKeyDb.CreatedAt,
		UpdatedAt:   privateKeyDb.UpdatedAt,
	}
}

// private key payload from PUT/POST
type privateKeyPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AlgorithmValue string `json:"algorithm.value"`
	PemContent     string `json:"pem"`
}

// new private key options
// used to return info about valid options when making a new key
type newPrivateKeyOptions struct {
	KeyAlgorithms []algorithm `json:"key_algorithms"`
}
