package download

import (
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"errors"
	"time"
)

// getKey returns the private key if the apiKey matches
// the requested key. It also checks the apiKeyViaUrl property if
// the client is making a request with the apiKey in the Url.
func (service *Service) getKey(keyName string, apiKey string, apiKeyViaUrl bool) (private_keys.Key, *output.JsonError) {
	// if apiKey is blank, definitely unauthorized
	if apiKey == "" {
		service.logger.Debug(errBlankApiKey)
		return private_keys.Key{}, output.JsonErrUnauthorized
	}

	// get the key from storage
	key, err := service.storage.GetOneKeyByName(keyName)
	if err != nil {
		// special error case for no record found
		if errors.Is(err, storage.ErrNoRecord) {
			service.logger.Debug(err)
			// exclude specific error since not authenticated
			return private_keys.Key{}, output.JsonErrNotFound(nil)
		} else {
			service.logger.Error(err)
			// exclude specific error since not authenticated
			return private_keys.Key{}, output.JsonErrStorageGeneric(nil)
		}
	}

	// if key is disabled via API, error
	if key.ApiKeyDisabled {
		service.logger.Debug(errApiDisabled)
		return private_keys.Key{}, output.JsonErrUnauthorized
	}

	// if apiKey came from URL, and key does not support this, error
	if apiKeyViaUrl && !key.ApiKeyViaUrl {
		service.logger.Debug(errApiKeyFromUrlDisallowed)
		return private_keys.Key{}, output.JsonErrUnauthorized
	}

	// verify apikey matches private key's apiKey (new or old)
	if (apiKey != key.ApiKey) && (apiKey != key.ApiKeyNew) {
		service.logger.Debug(errWrongApiKey)
		return private_keys.Key{}, output.JsonErrUnauthorized
	}

	// before return, update key last access, dont fail our though if this step fails, just log error
	err = service.storage.PutKeyLastAccess(key.ID, time.Now().Unix())
	if err != nil {
		service.logger.Errorf("download: failed to update key (id: %d) last access time (%s)", key.ID, err)
	}

	// return key
	return key, nil
}
