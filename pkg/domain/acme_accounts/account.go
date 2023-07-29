package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
)

// Account is a single ACME account
type Account struct {
	ID          int
	Name        string
	Description string
	AcmeServer  acme_servers.Server
	AccountKey  private_keys.Key
	Status      string
	Email       string
	AcceptedTos bool
	CreatedAt   int
	UpdatedAt   int
	Kid         string
}

// AccountSummaryResponse is a JSON response containing only
// fields desired for the summary
type AccountSummaryResponse struct {
	ID          int                          `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	AcmeServer  AccountServerSummaryResponse `json:"acme_server"`
	AccountKey  AccountKeySummaryResponse    `json:"private_key"`
	Status      string                       `json:"status"`
	Email       string                       `json:"email"`
	AcceptedTos bool                         `json:"accepted_tos"`
	IsStaging   bool                         `json:"is_staging"`
}

type AccountServerSummaryResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DirectoryURL string `json:"directory_url"`
	IsStaging    bool   `json:"is_staging"`
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
		AcmeServer: AccountServerSummaryResponse{
			ID:           acct.AcmeServer.ID,
			Name:         acct.AcmeServer.Name,
			DirectoryURL: acct.AcmeServer.DirectoryURL,
			IsStaging:    acct.AcmeServer.IsStaging,
		},
		AccountKey: AccountKeySummaryResponse{
			ID:   acct.AccountKey.ID,
			Name: acct.AccountKey.Name,
		},
		Status:      acct.Status,
		Email:       acct.Email,
		AcceptedTos: acct.AcceptedTos,
	}
}

// accountDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type accountDetailedResponse struct {
	AccountSummaryResponse
	AcmeServer accountServerDetailedResponse `json:"acme_server"`
	AccountKey accountKeyDetailedResponse    `json:"private_key"`
	CreatedAt  int                           `json:"created_at"`
	UpdatedAt  int                           `json:"updated_at"`
	Kid        string                        `json:"kid"`
}

type accountServerDetailedResponse struct {
	AccountServerSummaryResponse
	// from remote server
	ExternalAccountRequired bool   `json:"external_account_required"`
	TermsOfService          string `json:"terms_of_service"`
}

type accountKeyDetailedResponse struct {
	AccountKeySummaryResponse
	Algorithm key_crypto.Algorithm `json:"algorithm"`
}

func (acct Account) detailedResponse(service *Service) (accountDetailedResponse, error) {
	// get acme service for the account
	acmeService, err := service.acmeServerService.AcmeService(acct.AcmeServer.ID)
	if err != nil {
		return accountDetailedResponse{}, err
	}

	return accountDetailedResponse{
		AccountSummaryResponse: acct.SummaryResponse(),
		AcmeServer: accountServerDetailedResponse{
			AccountServerSummaryResponse: AccountServerSummaryResponse{
				ID:           acct.AcmeServer.ID,
				Name:         acct.AcmeServer.Name,
				DirectoryURL: acct.AcmeServer.DirectoryURL,
				IsStaging:    acct.AcmeServer.IsStaging,
			},
			ExternalAccountRequired: acmeService.RequiresEAB(),
			TermsOfService:          acmeService.TosUrl(),
		},
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
	}, nil
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
	AcmeServers   []acme_servers.ServerSummaryResponse `json:"acme_servers"`
	AvailableKeys []private_keys.KeySummaryResponse    `json:"private_keys"`
}
