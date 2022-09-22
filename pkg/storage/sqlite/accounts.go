package sqlite

import (
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// accountDb is a single acme account, as database table fields
// corresponds to acme_accounts.Account
type accountDb struct {
	id           int
	name         string
	description  string
	accountKeyDb accountKeyDb
	status       string
	email        string
	acceptedTos  bool
	isStaging    bool
}

type accountKeyDb struct {
	id   int
	name string
}

func (acct accountDb) toAccount() acme_accounts.Account {
	return acme_accounts.Account{
		ID:          acct.id,
		Name:        acct.name,
		Description: acct.description,
		AccountKey:  acct.accountKeyDb.toAccountKey(),
		Status:      acct.status,
		Email:       acct.email,
		AcceptedTos: acct.acceptedTos,
		IsStaging:   acct.isStaging,
	}
}

func (acctKey accountKeyDb) toAccountKey() acme_accounts.AccountKey {
	return acme_accounts.AccountKey{
		ID:   acctKey.id,
		Name: acctKey.name,
	}
}

// accountDbExtended is a single acme account, as database table
// fields. corresponds to acme_accounts.AccountExtended
type accountDbExtended struct {
	accountDb
	accountKeyDb accountKeyExtendedDb
	createdAt    int
	updatedAt    int
	kid          string
}

type accountKeyExtendedDb struct {
	accountKeyDb
	algorithmValue string
	pem            string
}

func (acct accountDbExtended) toAccountExtended() acme_accounts.AccountExtended {
	return acme_accounts.AccountExtended{
		// regular Key fields
		Account: acct.toAccount(),
		// extended fields
		AccountKey: acct.accountKeyDb.toAccountKeyExtended(),
		CreatedAt:  acct.createdAt,
		UpdatedAt:  acct.updatedAt,
		Kid:        acct.kid,
	}
}

func (acctKey accountKeyExtendedDb) toAccountKeyExtended() acme_accounts.AccountKeyExtended {
	return acme_accounts.AccountKeyExtended{
		AccountKey: acctKey.toAccountKey(),
		Algorithm:  key_crypto.AlgorithmByValue(acctKey.algorithmValue),
		Pem:        acctKey.pem,
	}
}
