package private_keys

import (
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
)

// idValid returns true if the specified keyId exists in storage
func (service *Service) idValid(keyId int) bool {
	// fetch key id from storage, if fails it doesn't exist
	_, err := service.storage.GetOneKeyById(keyId)

	// true if no error, no error if succesfully retrieved
	return err == nil
}

// nameValid returns true if the specified key name is acceptable and
// false if it is not. This check includes validating specified
// characters and also confirms the name is not already in use by another
// key. If an id is specified, the name will also be accepted if the name
// is already in use by the specified id.
func (service *Service) nameValid(keyName string, keyId *int) bool {
	// basic character/length check
	if !validation.NameValid(keyName) {
		return false
	}

	// make sure the name isn't already in use in storage
	key, err := service.storage.GetOneKeyByName(keyName)
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
