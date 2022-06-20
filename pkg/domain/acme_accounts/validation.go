package acme_accounts

import (
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/utils"
)

// isAccountIdExisting returns an error if not valid, nil if valid
func (service *Service) isIdExisting(idParam int, idPayload *int) error {
	// basic check
	err := utils.IsIdExistingMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	// check id exists in storage
	_, err = service.storage.GetOneAccountById(*idPayload)
	if err != nil {
		return err
	}

	return nil
}

// isNameValid returns an error if not valid, nil if valid
func (service *Service) isNameValid(idPayload *int, namePayload *string) error {
	// basic check
	err := utils.IsNameValid(namePayload)
	if err != nil {
		return err
	}

	// make sure the name isn't already in use in storage
	// the db
	account, err := service.storage.GetOneAccountByName(*namePayload)
	if err == sql.ErrNoRows {
		// no rows means name is not in use
		return nil
	} else if err != nil {
		// any other error, return the error
		return err
	}

	// if the returned account is the account being edited, no error
	if account.ID == *idPayload {
		return nil
	}

	return errors.New("name already in use")
}
