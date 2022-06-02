package acme_accounts

import (
	"legocerthub-backend/utils/acme_utils"
	"strings"
	"time"
)

// Turn LE response into Db object
func acmeResponseDbObj(accountId int, response acme_utils.AcmeAccountResponse) accountDb {
	var account accountDb

	account.id = accountId

	account.email.Valid = true
	account.email.String = strings.TrimPrefix(response.Contact[0], "mailto:")

	unixCreated, err := acme_utils.LeToUnixTime(response.CreatedAt)
	if err != nil {
		unixCreated = 0
	}
	account.createdAt = int(unixCreated)

	account.updatedAt = int(time.Now().Unix())

	account.status.Valid = true
	account.status.String = response.Status

	account.kid.Valid = true
	account.kid.String = response.Location

	return account
}

// Create account with LE
func (acme *AccountAppAcme) createLeAccount(payload accountPayload, keyPem string) (acme_utils.AcmeAccountResponse, error) {
	// payload to sent to LE
	var acmeAccount acme_utils.AcmeAccount
	acmeAccount.Contact = []string{"mailto:" + payload.Email}
	acmeAccount.TermsOfServiceAgreed = true

	// vars for return
	var acmeAccountResponse acme_utils.AcmeAccountResponse
	var err error

	if (payload.IsStaging == "true") || (payload.IsStaging == "on") {
		acmeAccountResponse, err = acme.StagingDir.CreateAccount(acmeAccount, keyPem)
		if err != nil {
			return acmeAccountResponse, err
		}
	} else {
		acmeAccountResponse, err = acme.ProdDir.CreateAccount(acmeAccount, keyPem)
		if err != nil {
			return acmeAccountResponse, err
		}
	}

	return acmeAccountResponse, nil
}
