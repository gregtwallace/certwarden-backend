package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// Account is a single ACME account
type Account struct {
	ID          int
	Name        string
	Description string
	AccountKey  private_keys.Key
	Status      string
	Email       string
	AcceptedTos bool
	IsStaging   bool
	CreatedAt   int
	UpdatedAt   int
	Kid         string
}

// AccountSummaryResponse is a JSON response containing only
// fields desired for the summary
type AccountSummaryResponse struct {
	ID          int                       `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	AccountKey  AccountKeySummaryResponse `json:"private_key"`
	Status      string                    `json:"status"`
	Email       string                    `json:"email"`
	AcceptedTos bool                      `json:"accepted_tos"`
	IsStaging   bool                      `json:"is_staging"`
}

type AccountKeySummaryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (acct Account) SummaryResponse() AccountSummaryResponse {
	return AccountSummaryResponse{
		ID:          acct.ID,
		Name:        acct.Name,
		Description: acct.Description,
		AccountKey: AccountKeySummaryResponse{
			ID:   acct.AccountKey.ID,
			Name: acct.AccountKey.Name,
		},
		Status:      acct.Status,
		Email:       acct.Email,
		AcceptedTos: acct.AcceptedTos,
		IsStaging:   acct.IsStaging,
	}
}

// accountDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type accountDetailedResponse struct {
	AccountSummaryResponse
	AccountKey accountKeyDetailedResponse `json:"private_key"`
	CreatedAt  int                        `json:"created_at"`
	UpdatedAt  int                        `json:"updated_at"`
	Kid        string                     `json:"kid"`
}

type accountKeyDetailedResponse struct {
	AccountKeySummaryResponse
	Algorithm key_crypto.Algorithm `json:"algorithm"`
}

func (acct Account) detailedResponse() accountDetailedResponse {
	return accountDetailedResponse{
		AccountSummaryResponse: acct.SummaryResponse(),
		AccountKey: accountKeyDetailedResponse{
			AccountKeySummaryResponse: AccountKeySummaryResponse{
				ID:   acct.AccountKey.ID,
				Name: acct.AccountKey.Name,
			},
			Algorithm: acct.AccountKey.Algorithm,
		},
		CreatedAt: acct.CreatedAt,
		UpdatedAt: acct.UpdatedAt,
		Kid:       acct.Kid,
	}
}

// AcmeAccountKey() provides a method to create an ACME AccountKey
// for the Account
func (account *Account) AcmeAccountKey() (acmeAcctKey acme.AccountKey, err error) {
	// get crypto key from the account's key
	acmeAcctKey.Key, err = account.AccountKey.CryptoPrivateKey()
	if err != nil {
		return acme.AccountKey{}, err
	}

	// set Kid from account
	acmeAcctKey.Kid = account.Kid

	return acmeAcctKey, nil
}

// newAccountPayload() generates the payload for ACME to post to the
// new-account endpoint
func (account *Account) newAccountPayload(eabKid string, eabHmacKey string) acme.NewAccountPayload {
	return acme.NewAccountPayload{
		TosAgreed:                     account.AcceptedTos,
		Contact:                       emailToContact(account.Email),
		ExternalAccountBindingKid:     eabKid,
		ExternalAccountBindingHmacKey: eabHmacKey,
	}
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	TosUrl        string                            `json:"tos_url"`
	StagingTosUrl string                            `json:"staging_tos_url"`
	AvailableKeys []private_keys.KeySummaryResponse `json:"private_keys"`
}
