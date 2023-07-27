package acme_servers

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteServer deletes an acme server from storage and terminates the
// related service.
func (service *Service) DeleteServer(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// verify server exists
	_, err = service.getServer(id)
	if err != nil {
		return err
	}

	// do not allow delete if there are any accounts using the server
	if service.storage.ServerHasAccounts(id) {
		service.logger.Warn("cannot delete server (in use)")
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

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "deleted",
		ID:      id,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
