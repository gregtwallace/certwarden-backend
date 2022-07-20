package certificates

import (
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

// isIdExisting returns an error if not valid, nil if valid
func (service *Service) isIdExisting(id int) (err error) {
	_, err = service.storage.GetOneCertById(id)
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
	account, err := service.storage.GetOneCertByName(*namePayload)
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
