package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

var (
	ErrIdBad   = errors.New("key id is invalid")
	ErrNameBad = errors.New("private key name is not valid")

	ErrApiKeyBad    = errors.New("api key is not valid (must be at least 10 chars in length)")
	ErrApiKeyNewBad = errors.New("api key (new) is not valid (must be at least 10 chars in length)")

	ErrKeyOptionNone     = errors.New("no key option method specified")
	ErrKeyOptionMultiple = errors.New("multiple key option methods specified")
)

// getKey returns the Key for the specified id or an
// error.
func (service *Service) getKey(id int) (Key, error) {
	// basic check
	if !validation.IsIdExistingValidRange(id) {
		service.logger.Debug(ErrIdBad)
		return Key{}, output.ErrValidationFailed
	}

	// get the key from storage
	key, _, err := service.storage.GetOneKeyById(id)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return Key{}, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return Key{}, output.ErrStorageGeneric
		}
	}

	return key, nil
}

// nameValid returns true if the specified key name is acceptable and
// false if it is not. This check includes validating specified
// characters and also confirms the name is not already in use by another
// key. If an id is specified, the name will also be accepted if the name
// is already in use by the specified id.
func (service *Service) NameValid(keyName string, keyId *int) bool {
	// basic character/length check
	if !validation.NameValid(keyName) {
		return false
	}

	// make sure the name isn't already in use in storage
	key, _, err := service.storage.GetOneKeyByName(keyName)
	if err == storage.ErrNoRecord {
		// no rows means name is not in use
		return true
	} else if err != nil {
		// any other error
		return false
	}

	// if the returned key is the key being edited, name is ok
	if keyId != nil && key.ID == *keyId {
		return true
	}

	return false
}

// GetAvailableKeys returns a list of all available keys; storage should
// return keys that exist but are not already in use by an account or a
// certificate
// TODO: Maybe move business logic here instead of in storage
func (service *Service) AvailableKeys() (keys []Key, err error) {
	return service.storage.GetAvailableKeys()
}

// KeyAvailable returns true if the specified keyId is available for
// use (i.e. not already in use by an account or a certificate)
func (service *Service) KeyAvailable(keyId int) bool {
	// get available keys list
	keys, err := service.AvailableKeys()
	if err != nil {
		return false
	}

	// verify specified key id is in the available list
	for i := range keys {
		if keys[i].ID == keyId {
			return true
		}
	}

	return false
}
