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
func (service *Service) DeleteServer(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// verify server exists
	_, outErr := service.getServer(id)
	if outErr != nil {
		return outErr
	}

	// do not allow delete if there are any accounts using the server
	if service.storage.ServerHasAccounts(id) {
		service.logger.Debug("cannot delete server (in use)")
		return output.ErrDeleteInUse
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteServer(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
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
		return output.ErrWriteJsonError
	}

	return nil
}
