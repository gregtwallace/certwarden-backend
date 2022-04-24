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

// a single private key
type privateKey struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description,omitempty"`
	Algorithm   string         `json:"algorithm"`
	Pem         string         `json:"pem"`
	ApiKey      string         `json:"api_key"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
