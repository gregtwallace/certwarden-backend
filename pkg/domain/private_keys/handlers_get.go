package private_keys

import (
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllKeys returns all of the private keys in storage as JSON
func (service *Service) GetAllKeys(w http.ResponseWriter, r *http.Request) (err error) {
	// get keys from storage
	keys, err := service.storage.GetAllKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, keys, "private_keys")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetOneKey returns a single private key as JSON
func (service *Service) GetOneKey(w http.ResponseWriter, r *http.Request) (err error) {
	// get key id and convert to int
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	err = validation.IsIdNew(&id)
	if err == nil {
		return service.GetNewKeyOptions(w, r)
	}

	// if id < 0 & not new, it is definitely not valid
	err = validation.IsIdExisting(&id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get the key from storage
	key, err := service.storage.GetOneKeyById(id)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, key, "private_key")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewKeyOptions returns configuration options for a new private key as JSON
func (service *Service) GetNewKeyOptions(w http.ResponseWriter, r *http.Request) error {
	newKeyOptions := newKeyOptions{}
	newKeyOptions.KeyAlgorithms = key_crypto.ListOfAlgorithms()

	// return response to client
	_, err := service.output.WriteJSON(w, http.StatusOK, newKeyOptions, "private_key_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// ServeKeyPem returns the private key to the client
func (service *Service) GetKeyPemFile(w http.ResponseWriter, r *http.Request) (err error) {
	// get key name
	keyName := httprouter.ParamsFromContext(r.Context()).ByName("name")

	// get api key from header
	apiKey := r.Header.Get("X-API-Key")
	// try to get from apikey header if X-API-Key was empty
	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}
	// if apiKey is blank, definitely not authorized
	if apiKey == "" {
		service.logger.Debug(err)
		return output.ErrUnauthorized
	}

	// get the key from storage
	key, err := service.storage.GetOneKeyByName(keyName)
	if err != nil {
		// special error case for no record found
		if err == storage.ErrNoRecord {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// verify apikey matches private key
	if apiKey != *key.ApiKey {
		service.logger.Debug(err)
		return output.ErrUnauthorized
	}

	// return pem file to client
	_, err = service.output.WritePem(w, *key.Pem)
	if err != nil {
		service.logger.Error(err)
		return output.ErrWritePemFailed
	}

	return nil
}
