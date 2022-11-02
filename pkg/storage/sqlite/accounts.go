package sqlite

import (
	"legocerthub-backend/pkg/domain/acme_accounts"
)

// accountDb is a single acme account, as database table fields
// corresponds to acme_accounts.Account
type accountDb struct {
	id           int
	name         string
	description  string
	accountKeyDb keyDb
	status       string
	email        string
	acceptedTos  bool
	isStaging    bool
	createdAt    int
	updatedAt    int
	kid          string
}

func (acct accountDb) toAccount() acme_accounts.Account {
	return acme_accounts.Account{
		ID:          acct.id,
		Name:        acct.name,
		Description: acct.description,
		AccountKey:  acct.accountKeyDb.toKey(),
		Status:      acct.status,
		Email:       acct.email,
		AcceptedTos: acct.acceptedTos,
		IsStaging:   acct.isStaging,
		CreatedAt:   acct.createdAt,
		UpdatedAt:   acct.updatedAt,
		Kid:         acct.kid,
	}
}
