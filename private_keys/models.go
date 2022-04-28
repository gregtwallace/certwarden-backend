package private_keys

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

// PrivateKeys struct for database access
type PrivateKeysApp struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
}

// a single private key (as returned from the db query)
type sqlPrivateKey struct {
	id          int
	name        string
	description sql.NullString
	algorithmId int
	pem         string
	apiKey      string
	createdAt   int
	updatedAt   int
}

// type to hold key algorithms
type algorithm struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// define the supported algorithms
// the id MUST be unique
// TODO: write a go test to confirm uniqueness
func supportedKeyAlgorithms() []algorithm {
	return []algorithm{
		{
			ID:   0,
			Name: "RSA 2048",
		},
		{
			ID:   1,
			Name: "ECDSA P-256",
		},
	}
}

// return an algorithm based on its ID
// if not found, return an error
func keyAlgorithmById(id int) (algorithm, error) {
	supportedAlgorithms := supportedKeyAlgorithms()

	for i := 0; i < len(supportedAlgorithms); i++ {
		if id == supportedAlgorithms[i].ID {
			return supportedAlgorithms[i], nil
		}
	}

	return algorithm{}, errors.New("privatekeys: algorithmById: invalid algorithm id")
}

// a single private key (suitable for the API)
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

// translate the db fetch into the api object
func (sqlPrivateKey *sqlPrivateKey) sqlToPrivateKey() (*privateKey, error) {
	keyAlgorithm, err := keyAlgorithmById(sqlPrivateKey.algorithmId)
	if err != nil {
		return nil, err
	}

	return &privateKey{
		ID:          sqlPrivateKey.id,
		Name:        sqlPrivateKey.name,
		Description: sqlPrivateKey.description.String,
		Algorithm:   keyAlgorithm,
		Pem:         sqlPrivateKey.pem,
		ApiKey:      sqlPrivateKey.apiKey,
		CreatedAt:   sqlPrivateKey.createdAt,
		UpdatedAt:   sqlPrivateKey.updatedAt,
	}, nil
}
