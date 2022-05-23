package acme_accounts

import (
	"database/sql"
	"log"
	"time"
)

// AcmeAccounts struct for database access
type AcmeAccountsApp struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
}

// a single acmeAccount
type acmeAccount struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	PrivateKeyID   int    `json:"private_key_id"`
	PrivateKeyName string `json:"private_key_name"` // comes from a join with key table
	Description    string `json:"description"`
	Status         string `json:"status"`
	Email          string `json:"email"`
	AcceptedTos    bool   `json:"accepted_tos,omitempty"`
	IsStaging      bool   `json:"is_staging"`
	CreatedAt      int    `json:"created_at,omitempty"`
	UpdatedAt      int    `json:"updated_at,omitempty"`
	Kid            string `json:"kid,omitempty"`
}

// a single private key, as database table fields
type acmeAccountDb struct {
	id             int
	name           string
	privateKeyId   int
	privateKeyName sql.NullString // comes from a join with key table
	description    sql.NullString
	status         sql.NullString
	email          sql.NullString
	acceptedTos    sql.NullBool
	isStaging      sql.NullBool
	createdAt      int
	updatedAt      int
	kid            sql.NullString
}

// translate the db object into the api object
func (acmeAccountDb *acmeAccountDb) acmeAccountDbToAcc() (*acmeAccount, error) {
	return &acmeAccount{
		ID:             acmeAccountDb.id,
		Name:           acmeAccountDb.name,
		PrivateKeyID:   acmeAccountDb.privateKeyId,
		PrivateKeyName: acmeAccountDb.privateKeyName.String,
		Description:    acmeAccountDb.description.String,
		Status:         acmeAccountDb.status.String,
		Email:          acmeAccountDb.email.String,
		AcceptedTos:    acmeAccountDb.acceptedTos.Bool,
		IsStaging:      acmeAccountDb.isStaging.Bool,
		CreatedAt:      acmeAccountDb.createdAt,
		UpdatedAt:      acmeAccountDb.updatedAt,
		Kid:            acmeAccountDb.kid.String,
	}, nil
}
