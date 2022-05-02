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

// Response backend sends in response to PUT/POST
type jsonResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// a single private key
type privateKey struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Algorithm   algorithm `json:"algorithm"`
	Pem         string    `json:"pem"`
	ApiKey      string    `json:"api_key"`
	CreatedAt   int       `json:"created_at"`
	UpdatedAt   int       `json:"updated_at"`
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
func (sqlPrivateKey *privateKeyDb) privateKeyDbToPk() (*privateKey, error) {
	return &privateKey{
		ID:          sqlPrivateKey.id,
		Name:        sqlPrivateKey.name,
		Description: sqlPrivateKey.description.String,
		Algorithm:   algorithmByValue(sqlPrivateKey.algorithmValue),
		Pem:         sqlPrivateKey.pem,
		ApiKey:      sqlPrivateKey.apiKey,
		CreatedAt:   sqlPrivateKey.createdAt,
		UpdatedAt:   sqlPrivateKey.updatedAt,
	}, nil
}

// private key payload from PUT/POST
type privateKeyPayload struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	AlgorithmValue string `json:"algorithm.value"`
}
