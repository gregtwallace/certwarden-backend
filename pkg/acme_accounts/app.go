package acme_accounts

import (
	"legocerthub-backend/pkg/utils/acme_utils"
	"strconv"
)

// createNewAccount creates a new account and registers it with LE
func (app *AccountsApp) createNewAccount(payload accountPayload) error {
	// load fields
	account, err := payload.accountPayloadToDb()
	if err != nil {
		return err
	}

	// post initial account
	payload.ID, err = app.DB.postNewAccount(account)
	if err != nil {
		return err
	}

	err = app.createNewLEAccount(payload)
	if err != nil {
		return err
	}

	return nil
}

// createNewLEAccount takes a payload and registers it with LE as a new account
func (app *AccountsApp) createNewLEAccount(payload accountPayload) error {
	var acmeAccountResponse acme_utils.AcmeAccountResponse

	// fetch appropriate key
	keyPem, err := app.DB.getAccountKeyPem(payload.ID)
	if err != nil {
		return err
	}

	acmeAccountResponse, err = app.Acme.createLeAccount(payload, keyPem)
	if err != nil {
		return err
	}

	// Write the returned account info from LE to the db
	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		return err
	}

	err = app.DB.putLEAccountInfo(acmeResponseDbObj(id, acmeAccountResponse))
	if err != nil {
		return err
	}

	return nil
}

// updateLEAccount updates account settings with LE
func (app *AccountsApp) updateLEAccount(payload accountPayload) error {
	var acmeAccountResponse acme_utils.AcmeAccountResponse

	// fetch appropriate key
	keyPem, err := app.DB.getAccountKeyPem(payload.ID)
	if err != nil {
		return err
	}

	// get kid
	kid, err := app.DB.getAccountKid(payload.ID)
	if err != nil {
		return err
	}

	acmeAccountResponse, err = app.Acme.updateLeAccount(payload, keyPem, kid)
	if err != nil {
		return err
	}

	// Write the returned account info from LE to the db
	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		return err
	}

	err = app.DB.putLEAccountInfo(acmeResponseDbObj(id, acmeAccountResponse))
	if err != nil {
		return err
	}

	return nil
}
