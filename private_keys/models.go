package private_keys

import (
	"database/sql"
	"log"
	"time"
)

// PrivateKeys struct for database access
type PrivateKeysDB struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
}

// a single private key (as returned from the db query)
type sqlPrivateKey struct {
	ID          int
	Name        string
	Description sql.NullString
	Algorithm   string
	Pem         string
	ApiKey      string
	CreatedAt   int
	UpdatedAt   int
}

// a single private key (suitable for the API)
type privateKey struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Algorithm   string `json:"algorithm,omitempty"`
	Pem         string `json:"pem,omitempty"`
	ApiKey      string `json:"api_key,omitempty"`
	CreatedAt   int    `json:"created_at,omitempty"`
	UpdatedAt   int    `json:"updated_at,omitempty"`
}

// translate the db fetch into the api object
func (sqlPrivateKey *sqlPrivateKey) sqlToPrivateKey() *privateKey {
	return &privateKey{
		ID:          sqlPrivateKey.ID,
		Name:        sqlPrivateKey.Name,
		Description: sqlPrivateKey.Description.String,
		Algorithm:   sqlPrivateKey.Algorithm,
		Pem:         sqlPrivateKey.Pem,
		ApiKey:      sqlPrivateKey.ApiKey,
		CreatedAt:   sqlPrivateKey.CreatedAt,
		UpdatedAt:   sqlPrivateKey.UpdatedAt,
	}
}
