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

	// make keysResponse (for json output)
	var keysResponse []KeySummaryResponse
	for i := range keys {
		keysResponse = append(keysResponse, keys[i].SummaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, keysResponse, "private_keys")
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
	if validation.IsIdNew(id) {
		return service.GetNewKeyOptions(w, r)
	}

	// if id < 0 & not new, it is definitely not valid
	if !validation.IsIdExisting(id) {
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
	_, err = service.output.WriteJSON(w, http.StatusOK, key.detailedResponse(service.https || service.devMode), "private_key")
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
