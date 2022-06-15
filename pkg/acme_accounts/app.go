package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/utils/acme_utils"
)

// createNewAccount creates a new account and registers it with LE
func (service *Service) createNewAccount(payload AccountPayload) error {
	var err error

	// post initial account
	if payload.ID == nil {
		// TODO: remove if checking nil pointer in validation
		return errors.New("payload id cannot be nil")
	}
	*payload.ID, err = service.storage.PostNewAccount(payload)
	if err != nil {
		return err
	}

	err = service.createNewLEAccount(payload)
	if err != nil {
		return err
	}

	return nil
}

// createNewLEAccount takes a payload and registers it with LE as a new account
func (service *Service) createNewLEAccount(payload AccountPayload) error {
	var acmeAccountResponse acme_utils.AcmeAccountResponse

	// fetch appropriate key
	keyPem, err := service.storage.GetAccountPem(*payload.ID)
	if err != nil {
		return err
	}

	acmeAccountResponse, err = service.createLeAccount(payload, keyPem)
	if err != nil {
		return err
	}

	// Write the returned account info from LE to the db
	err = service.storage.PutLEAccountInfo(*payload.ID, acmeAccountResponse)
	if err != nil {
		return err
	}

	return nil
}

// updateLEAccount updates account settings with LE
func (service *Service) updateLEAccount(payload AccountPayload) error {
	var acmeAccountResponse acme_utils.AcmeAccountResponse

	// fetch appropriate key
	keyPem, err := service.storage.GetAccountPem(*payload.ID)
	if err != nil {
		return err
	}

	// get kid
	kid, err := service.storage.GetAccountKid(*payload.ID)
	if err != nil {
		return err
	}

	acmeAccountResponse, err = service.updateLeAccount(payload, keyPem, kid)
	if err != nil {
		return err
	}

	// Write the returned account info from LE to the db
	err = service.storage.PutLEAccountInfo(*payload.ID, acmeAccountResponse)
	if err != nil {
		return err
	}

	return nil
}
