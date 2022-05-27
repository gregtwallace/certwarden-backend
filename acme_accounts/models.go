package acme_accounts

import (
	"database/sql"
	"legocerthub-backend/private_keys"
	"legocerthub-backend/utils/acme_utils"
	"log"
	"time"
)

// AcmeAccounts struct for database access
type AcmeAccountsApp struct {
	Database *sql.DB
	Timeout  time.Duration
	Logger   *log.Logger
	Acme     *acme_utils.Acme
}

// a single acmeAccount
type acmeAccount struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PrivateKeyID   int    `json:"private_key_id"`
	PrivateKeyName string `json:"private_key_name"` // comes from a join with key table
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
	description    sql.NullString
	privateKeyId   sql.NullInt32
	privateKeyName sql.NullString // comes from a join with key table
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
		Description:    acmeAccountDb.description.String,
		PrivateKeyID:   int(acmeAccountDb.privateKeyId.Int32),
		PrivateKeyName: acmeAccountDb.privateKeyName.String,
		Status:         acmeAccountDb.status.String,
		Email:          acmeAccountDb.email.String,
		AcceptedTos:    acmeAccountDb.acceptedTos.Bool,
		IsStaging:      acmeAccountDb.isStaging.Bool,
		CreatedAt:      acmeAccountDb.createdAt,
		UpdatedAt:      acmeAccountDb.updatedAt,
		Kid:            acmeAccountDb.kid.String,
	}, nil
}

// acme account payload from PUT/POST
type acmeAccountPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	AcceptedTos string `json:"accepted_tos"`
}

// new account info
// used to return info about valid options when making a new account
type NewAcmeAccountOptions struct {
	TosUrl        string                    `json:"tos_url"`
	StagingTosUrl string                    `json:"staging_tos_url"`
	AvailableKeys []private_keys.PrivateKey `json:"available_keys"`
}
