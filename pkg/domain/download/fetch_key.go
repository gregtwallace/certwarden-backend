package download

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"errors"
)

// getKey returns the private key if the apiKey matches
// the requested key. It also checks the apiKeyViaUrl property if
// the client is making a request with the apiKey in the Url.
func (service *Service) getKey(keyName string, apiKey string, apiKeyViaUrl bool) (private_keys.Key, *output.Error) {
	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return private_keys.Key{}, output.ErrUnauthorized
	}

	// get the key from storage
	key, err := service.storage.GetOneKeyByName(keyName)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			return private_keys.Key{}, output.ErrNotFound
		} else {
			service.logger.Error(err)
			return private_keys.Key{}, output.ErrStorageGeneric
		}
	}

	// if key is disabled via API, error
	if key.ApiKeyDisabled {
		service.logger.Debug(errApiDisabled)
		return private_keys.Key{}, output.ErrUnauthorized
	}

	// if apiKey came from URL, and key does not support this, error
	if apiKeyViaUrl && !key.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return private_keys.Key{}, output.ErrUnauthorized
	}

	// verify apikey matches private key's apiKey (new or old)
	if (apiKey != key.ApiKey) && (apiKey != key.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return private_keys.Key{}, output.ErrUnauthorized
	}

	// return key
	return key, nil
}
