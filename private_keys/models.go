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
type privateKey struct {
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
type privateKeyDb struct {
	id             int
	name           string
	description    sql.NullString
	algorithmValue string
	pem            string
	apiKey         string
	createdAt      int
	updatedAt      int
}

// translate the db object into the api object
func (privateKeyDb *privateKeyDb) privateKeyDbToPk() (*privateKey, error) {
	return &privateKey{
		ID:          privateKeyDb.id,
		Name:        privateKeyDb.name,
		Description: privateKeyDb.description.String,
		Algorithm:   algorithmByValue(privateKeyDb.algorithmValue),
		Pem:         privateKeyDb.pem,
		ApiKey:      privateKeyDb.apiKey,
		CreatedAt:   privateKeyDb.createdAt,
		UpdatedAt:   privateKeyDb.updatedAt,
	}, nil
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
type NewPrivateKeyOptions struct {
	KeyAlgorithms []algorithm `json:"key_algorithms"`
}
