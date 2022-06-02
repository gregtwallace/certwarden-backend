package acme_accounts

import (
	"legocerthub-backend/utils/acme_utils"
	"strconv"
)

// createNewAccount creates a new account and registers it with LE
func (app *AccountsApp) createNewAccount(payload accountPayload) error {
	// vars for return

	var err error

	// load fields
	account, err := payload.accountPayloadToDb()
	if err != nil {
		app.Logger.Printf("accounts: createLeAccount: failed to load payload into db obj -- err: %s", err)
		return err
	}

	// post initial account
	payload.ID, err = app.DB.postNewAccount(account)
	if err != nil {
		app.Logger.Printf("accounts: createLeAccount: failed to write db -- err: %s", err)
		return err
	}

	err = app.createNewLEAccount(payload)
	if err != nil {
		app.Logger.Printf("accounts: createLeAccount: failed to create new LE account -- err: %s", err)
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
		app.Logger.Printf("accounts: createLeAccount: failed to create LE account -- err: %s", err)
		return err
	}

	// Write the returned account info from LE to the db
	id, err := strconv.Atoi(payload.ID)
	if err != nil {
		return err
	}

	err = app.DB.putLEAccountInfo(acmeResponseDbObj(id, acmeAccountResponse))
	if err != nil {
		app.Logger.Printf("accounts: createLeAccount: failed to update db -- err: %s", err)
		return err
	}

	return nil
}
