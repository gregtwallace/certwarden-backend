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

// type to hold key algorithms
type algorithm struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// define the supported algorithms
// the Value must be unique
// This MUST be kept in sync with the front end list
// TODO: write a go test to confirm uniqueness
func supportedKeyAlgorithms() []algorithm {
	return []algorithm{
		{
			Value: "rsa2048",
			Name:  "RSA 2048",
		},
		{
			Value: "ecdsa256",
			Name:  "ECDSA P-256",
		},
	}
}

// return an algorithm based on its value
// if not found, return an error
func keyAlgorithmByValue(dbValue string) (algorithm, error) {
	supportedAlgorithms := supportedKeyAlgorithms()

	for i := 0; i < len(supportedAlgorithms); i++ {
		if dbValue == supportedAlgorithms[i].Value {
			return supportedAlgorithms[i], nil
		}
	}

	return algorithm{}, errors.New("privatekeys: algorithmByValue: invalid algorithm value")
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
func (sqlPrivateKey *privateKeyDb) privateKeyDbToPk() (*privateKey, error) {
	keyAlgorithm, err := keyAlgorithmByValue(sqlPrivateKey.algorithmValue)
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

// private key payload from PUT/POST
type privateKeyPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Response backend sends in response to PUT/POST
type jsonResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
