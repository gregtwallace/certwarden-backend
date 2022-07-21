package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"
)

// a single ACME Account
type Account struct {
	ID          *int              `json:"id,omitempty"`
	Name        *string           `json:"name,omitempty"`
	Description *string           `json:"description,omitempty"`
	PrivateKey  *private_keys.Key `json:"private_key,omitempty"`
	Status      *string           `json:"status,omitempty"`
	Email       *string           `json:"email,omitempty"`
	AcceptedTos *bool             `json:"accepted_tos,omitempty"`
	IsStaging   *bool             `json:"is_staging,omitempty"`
	CreatedAt   *int              `json:"created_at,omitempty"`
	UpdatedAt   *int              `json:"updated_at,omitempty"`
	Kid         *string           `json:"kid,omitempty"`
}

// AccountKey() returns the ACME AccountKey which is a combination of the
// crypto.PrivateKey and Kid
func (account *Account) AccountKey() (accountKey acme.AccountKey, err error) {
	accountKey.Key, err = account.PrivateKey.CryptoKey()
	if err != nil {
		return acme.AccountKey{}, err
	}

	// if kid is nil, make kid blank
	if account.Kid != nil {
		accountKey.Kid = *account.Kid
	} else {
		accountKey.Kid = ""
	}

	return accountKey, nil
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"private_keys"`
}
