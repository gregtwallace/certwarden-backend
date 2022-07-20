package private_keys

import (
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

// isIdExisting returns an error if not valid (ie the key does not exist)
func (service *Service) isIdExisting(id int) (err error) {
	// check id exists in storage
	_, err = service.storage.GetOneKeyById(id)
	if err != nil {
		return err
	}

	return nil
}

// IsIdExistingMatch returns an error if not valid, nil if valid
// we'll generally assume the id is valid if >= 0
func (service *Service) isIdExistingMatch(idParam int, idPayload *int) error {
	// basic check
	err := validation.IsIdExistingMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	// check id exists
	err = service.isIdExisting(*idPayload)
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
	account, err := service.storage.GetOneKeyByName(*namePayload)
	if err == storage.ErrNoRecord {
		// no rows means name is not in use
		return nil
	} else if err != nil {
		// any other error, return the error
		return err
	}

	// if the returned key is the key being edited, no error
	if *account.ID == *idPayload {
		return nil
	}

	return validation.ErrNameInUse
}

// GetAvailableKeys returns a list of all available keys; storage should
// return keys that exist but are not already in use by an account or a
// certificate
func (service *Service) GetAvailableKeys() (keys []Key, err error) {
	return service.storage.GetAvailableKeys()
}

// IsPrivateKeyValid returns an error if the key is not valid and available
func (service *Service) IsPrivateKeyAvailable(keyId *int) error {
	// get available keys list
	keys, err := service.GetAvailableKeys()
	if err != nil {
		return err
	}

	// verify specified key id is in the available list
	for _, key := range keys {
		if *key.ID == *keyId {
			return nil
		}
	}

	return validation.ErrKeyBad
}
