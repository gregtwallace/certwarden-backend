package acme_accounts

import (
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

// isIdExisting returns an error if not valid, nil if valid
func (service *Service) isIdExisting(id int) (err error) {
	_, err = service.storage.GetOneAccountById(id, false)
	if err != nil {
		return err
	}

	return nil
}

// isIdExisting returns an error if not valid, nil if valid
func (service *Service) isIdExistingMatch(idParam int, idPayload *int) error {
	// basic check
	err := validation.IsIdExistingMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	// check id exists in storage
	err = service.isIdExisting(idParam)
	if err != nil {
		return err
	}

	return nil
}

// isNameValid returns an error if not valid, nil if valid
func (service *Service) isNameValid(idPayload *int, namePayload *string) error {
	// basic check
	err := validation.IsNameValid(namePayload)
	if err != nil {
		return err
	}

	// make sure the name isn't already in use in storage
	// the db
	account, err := service.storage.GetOneAccountByName(*namePayload, false)
	if err == storage.ErrNoRecord {
		// no rows means name is not in use
		return nil
	} else if err != nil {
		// any other error, return the error
		return err
	}

	// if the returned account is the account being edited, no error
	if *account.ID == *idPayload {
		return nil
	}

	return validation.ErrNameInUse
}

// GetAvailableAccounts returns a list of accounts that have status == valid and have also
// accepted the ToS (which is probably redundant)
func (service *Service) GetAvailableAccounts() ([]Account, error) {
	return service.storage.GetAvailableAccounts()
}

// IsAcmeAccountValid returns an error if the account does not exist or does not have
// a status == valid, and accepted_tos == true
func (service *Service) IsAcmeAccountValid(accountId *int) error {
	// get available accounts list
	accounts, err := service.GetAvailableAccounts()
	if err != nil {
		return err
	}

	// verify specified account id is available
	for _, account := range accounts {
		if *account.ID == *accountId {
			return nil
		}
	}

	return validation.ErrKeyBad
}
