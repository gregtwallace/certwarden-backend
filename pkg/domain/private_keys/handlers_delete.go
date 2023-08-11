package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// DeleteKey deletes a private key from storage
func (service *Service) DeleteKey(w http.ResponseWriter, r *http.Request) (err error) {
	// get params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validate key exists
	_, err = service.getKey(id)
	if err != nil {
		return err
	}

	// confirm key is not in use
	inUse, err := service.storage.KeyInUse(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	if inUse {
		service.logger.Warn("cannot delete, in use")
		return output.ErrDeleteInUse
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteKey(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "deleted",
		ID:      id,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}

// RemoveOldApiKey discards a key's api_key, replaces it with the key's
// api_key_new, and then blanks api_key_new
func (service *Service) RemoveOldApiKey(w http.ResponseWriter, r *http.Request) (err error) {
	// get id param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	keyId, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// get key (validate exists)
	key, err := service.getKey(keyId)
	if err != nil {
		return err
	}

	// verify new api key is not empty (need something to promote)
	if key.ApiKeyNew == "" {
		service.logger.Debug(errors.New("new api key does not exist"))
		return output.ErrValidationFailed
	}
	// validation -- end

	// update storage
	// set current api key from new key
	err = service.storage.PutKeyApiKey(keyId, key.ApiKeyNew, int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	// set new key to blank
	err = service.storage.PutKeyNewApiKey(keyId, "", int(time.Now().Unix()))
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "new api key promoted", // TODO?
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
