package acme_accounts

import (
	"legocerthub-backend/utils/acme_utils"
)

func (app *AccountsApp) createLeAccount(payload accountPayload) (acme_utils.AcmeAccountResponse, error) {
	// vars for return
	var acmeAccountResponse acme_utils.AcmeAccountResponse
	var err error

	// fetch appropriate key
	keyPem, err := app.DB.getKeyPem(payload.PrivateKeyID)
	if err != nil {
		return acmeAccountResponse, err
	}

	return app.Acme.createLeAccount(payload, keyPem)
}
