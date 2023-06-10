package sqlite

import (
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// accountDb is a single acme account, as database table fields
// corresponds to acme_accounts.Account
type accountDb struct {
	id              int
	name            string
	description     string
	accountServerDb acmeServerDb
	accountKeyDb    keyDb
	status          string
	email           string
	acceptedTos     bool
	createdAt       int
	updatedAt       int
	kid             string
}

func (acct accountDb) toAccount() acme_accounts.Account {
	return acme_accounts.Account{
		ID:          acct.id,
		Name:        acct.name,
		Description: acct.description,
		AcmeServer:  acct.accountServerDb.toServer(),
		AccountKey:  acct.accountKeyDb.toKey(),
		Status:      acct.status,
		Email:       acct.email,
		AcceptedTos: acct.acceptedTos,
		CreatedAt:   acct.createdAt,
		UpdatedAt:   acct.updatedAt,
		Kid:         acct.kid,
	}
}
