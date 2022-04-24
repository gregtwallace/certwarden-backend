package acme_accounts

import (
	"database/sql"
	"log"
	"time"
)

// AcmeAccounts struct for database access
type AcmeAccountsDB struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
}

// a single acmeAccount
type acmeAccount struct {
	ID             int       `json:"id"`
	PrivateKeyID   int       `json:"private_key_id"`
	PrivateKeyName string    `json:"private_key_name"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Description    string    `json:"description"`
	AcceptedTos    bool      `json:"accepted_tos"`
	IsStaging      bool      `json:"is_staging"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
