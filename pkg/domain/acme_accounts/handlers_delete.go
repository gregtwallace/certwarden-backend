package acme_accounts

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteAccount deletes an acme account from storage
func (service *Service) DeleteAccount(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// verify account exists
	_, err = service.getAccount(id)
	if err != nil {
		return err
	}

	// do not allow delete if there are any certs using the account
	if service.storage.AccountHasCerts(id) {
		service.logger.Warn("cannot delete account (in use)")
		return output.ErrDeleteInUse
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteAccount(id)
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
		return err
	}

	return nil
}
