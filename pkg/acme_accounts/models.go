package acme_accounts

import (
	"database/sql"
	"legocerthub-backend/pkg/private_keys"
	"legocerthub-backend/pkg/utils/acme_utils"
	"log"
	"strconv"
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
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Email        string `json:"email"`
	PrivateKeyID string `json:"private_key_id"`
	AcceptedTos  string `json:"accepted_tos"`
	IsStaging    string `json:"is_staging"`
}

// new account info
// used to return info about valid options when making a new account
type NewAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"available_keys"`
}

// load account payload into a db object
func (payload *accountPayload) accountPayloadToDb() (accountDb, error) {
	var accountDb accountDb

	accountId, err := strconv.Atoi(payload.ID)
	if err != nil {
		return accountDb, err
	}
	accountDb.id = accountId

	accountDb.name = payload.Name

	accountDb.description.Valid = true
	accountDb.description.String = payload.Description

	if payload.PrivateKeyID != "" {
		keyId, err := strconv.Atoi(payload.PrivateKeyID)
		if err != nil {
			return accountDb, err
		}
		accountDb.privateKeyId.Valid = true
		accountDb.privateKeyId.Int32 = int32(keyId)
	} else {
		accountDb.privateKeyId.Valid = false
	}

	accountDb.status.Valid = true
	accountDb.status.String = "Unknown"

	accountDb.email.Valid = true
	accountDb.email.String = payload.Email

	accountDb.acceptedTos.Valid = true
	accountDb.acceptedTos.Bool = true

	accountDb.isStaging.Valid = true
	if (payload.IsStaging == "true") || (payload.IsStaging == "on") {
		accountDb.isStaging.Bool = true
	} else {
		accountDb.isStaging.Bool = false
	}

	accountDb.createdAt = 0 // will update later from LE
	accountDb.updatedAt = int(time.Now().Unix())

	return accountDb, nil
}
