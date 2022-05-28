package acme_accounts

import (
	"database/sql"
	"legocerthub-backend/private_keys"
	"legocerthub-backend/utils/acme_utils"
	"log"
	"time"
)

type AccountAppDb struct {
	Database *sql.DB
	Timeout  time.Duration
}

type AccountAppAcme struct {
	ProdDir    *acme_utils.AcmeDirectory
	StagingDir *acme_utils.AcmeDirectory
}

// AcmeAccounts struct for database access
type AccountsApp struct {
	Logger *log.Logger
	DB     AccountAppDb
	Acme   AccountAppAcme
}

// a single ACME Account
type account struct {
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

// a single account, as database table fields
type accountDb struct {
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
func (accountDb *accountDb) accountDbToAcc() (account, error) {
	return account{
		ID:             accountDb.id,
		Name:           accountDb.name,
		Description:    accountDb.description.String,
		PrivateKeyID:   int(accountDb.privateKeyId.Int32),
		PrivateKeyName: accountDb.privateKeyName.String,
		Status:         accountDb.status.String,
		Email:          accountDb.email.String,
		AcceptedTos:    accountDb.acceptedTos.Bool,
		IsStaging:      accountDb.isStaging.Bool,
		CreatedAt:      accountDb.createdAt,
		UpdatedAt:      accountDb.updatedAt,
		Kid:            accountDb.kid.String,
	}, nil
}

// acme account payload from PUT/POST
type accountPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	AcceptedTos string `json:"accepted_tos"`
}

// new account info
// used to return info about valid options when making a new account
type NewAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"available_keys"`
}
