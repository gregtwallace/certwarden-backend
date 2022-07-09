package private_keys

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage"
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

	// TODO: Validate not in use, though storage should do this with foreign key check

	// delete from storage
	err = service.storage.DeleteKey(id)
	if err != nil {
		// special error case for in use
		if err == storage.ErrInUse {
			service.logger.Warn(err)
			return output.ErrDeleteInUse
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "deleted",
		ID:      id,
	}

	_, err = output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
