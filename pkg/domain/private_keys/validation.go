package private_keys

import (
	"database/sql"
	"errors"
	"legocerthub-backend/pkg/utils"
)

// IsIdExistingValid returns an error if not valid, nil if valid
// we'll generally assume the id is valid if >= 0
func (service *Service) isIdExisting(idParam int, idPayload *int) error {
	// basic check
	err := utils.IsIdExistingMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	// check id exists in storage
	_, err = service.storage.GetOneKeyById(*idPayload)
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
	account, err := service.storage.GetOneKeyByName(*namePayload)
	if err == sql.ErrNoRows {
		// no rows means name is not in use
		return nil
	} else if err != nil {
		// any other error, return the error
		return err
	}

	// if the returned key is the key being edited, no error
	if account.ID == *idPayload {
		return nil
	}

	return errors.New("name already in use")
}
