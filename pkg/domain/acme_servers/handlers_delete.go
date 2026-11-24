package acme_servers

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteServer deletes an acme server from storage and terminates the
// related service.
func (service *Service) DeleteServer(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// verify server exists
	_, outErr := service.getServer(id)
	if outErr != nil {
		return outErr
	}

	// do not allow delete if there are any accounts using the server
	inUse, err := service.storage.ServerInUse(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}
	if inUse {
		service.logger.Debug("cannot delete, in use")
		return output.ErrorJsonErrDeleteInUse("acme server")
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteServer(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	// delete acme Service
	service.mu.Lock()
	defer service.mu.Unlock()
	delete(service.acmeServers, id)

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted acme server (id: %d)", id),
	}

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
