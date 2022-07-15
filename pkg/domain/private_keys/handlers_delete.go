package private_keys

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteKey deletes a private key from storage
func (service *Service) DeleteKey(w http.ResponseWriter, r *http.Request) (err error) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// verify key exists
	err = service.isIdExisting(id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrNotFound
	}

	// confirm key is not in use
	inUse, err := service.storage.KeyInUse(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}
	if inUse == true {
		service.logger.Warn("cannot delete, in use")
		return output.ErrDeleteInUse
	}

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

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
