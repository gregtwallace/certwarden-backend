package certificates

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteCert deletes a cert from storage
func (service *Service) DeleteCert(w http.ResponseWriter, r *http.Request) (err error) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("certid")

	// convert id param to an integer
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// verify cert exists
	_, err = service.isIdExisting(id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrNotFound
	}

	// delete from storage
	err = service.storage.DeleteCert(id)
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
