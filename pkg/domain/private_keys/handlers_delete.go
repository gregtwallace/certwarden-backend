package private_keys

import (
	"certwarden-backend/pkg/output"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DeleteKey deletes a private key from storage
func (service *Service) DeleteKey(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validate key exists
	_, outErr := service.getKey(id)
	if outErr != nil {
		return outErr
	}

	// confirm key is not in use
	inUse, err := service.storage.KeyInUse(id)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}
	if inUse {
		service.logger.Warn("cannot delete, in use")
		return output.JsonErrDeleteInUse("private key")
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteKey(id)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted private key (id: %d)", id),
	}

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// RemoveOldApiKey discards a key's api_key, replaces it with the key's
// api_key_new, and then blanks api_key_new
func (service *Service) RemoveOldApiKey(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	keyId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// get key (validate exists)
	key, outErr := service.getKey(keyId)
	if outErr != nil {
		return outErr
	}

	// verify new api key is not empty (need something to promote)
	if key.ApiKeyNew == "" {
		err = errors.New("new api key does not exist")
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	// validation -- end

	// update storage
	// set current api key from new key
	err = service.storage.PutKeyApiKey(keyId, key.ApiKeyNew, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}
	key.ApiKey = key.ApiKeyNew

	// set new key to blank
	err = service.storage.PutKeyNewApiKey(keyId, "", int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}
	key.ApiKeyNew = ""

	// write response
	response := &privateKeyResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "private key old api key deleted, new api key promoted"
	response.PrivateKey = key.detailedResponse()

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
