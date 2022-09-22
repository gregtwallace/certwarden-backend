package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// Account is a single ACME account (summary)
type Account struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AccountKey  AccountKey `json:"private_key"`
	Status      string     `json:"status"`
	Email       string     `json:"email"`
	AcceptedTos bool       `json:"accepted_tos"`
	IsStaging   bool       `json:"is_staging"`
}

// AccountKey is a modified Private Key that only holds fields the
// account will use
type AccountKey struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// AccountExtended is a single ACME account with all of its details
type AccountExtended struct {
	Account
	AccountKey AccountKeyExtended `json:"private_key"`
	CreatedAt  int                `json:"created_at"`
	UpdatedAt  int                `json:"updated_at"`
	Kid        string             `json:"kid"`
}

// AccountKeyExtended is a modified Private Key that only holds fields
// the extended account will use
type AccountKeyExtended struct {
	AccountKey
	Algorithm key_crypto.Algorithm `json:"algorithm"`
	Pem       string               `json:"-"`
}

// AcmeAccountKey() returns the ACME acme.AccountKey which is a combination
// of the crypto.PrivateKey and Kid
func (account *AccountExtended) AcmeAccountKey() (acmeAccountKey acme.AccountKey, err error) {
	// generate key from account key pem
	acmeAccountKey.Key, err = key_crypto.PemStringToKey(account.AccountKey.Pem, account.AccountKey.Algorithm)
	if err != nil {
		return acme.AccountKey{}, err
	}

	// set Kid from account
	acmeAccountKey.Kid = account.Kid

	return acmeAccountKey, nil
}

// newAccountPayload() generates the payload for ACME to post to the
// new-account endpoint
func (account *Account) newAccountPayload() acme.NewAccountPayload {
	return acme.NewAccountPayload{
		TosAgreed: account.AcceptedTos,
		Contact:   emailToContact(account.Email),
	}
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"private_keys"`
}
