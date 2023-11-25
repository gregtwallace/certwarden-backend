package acme_accounts

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteAccount deletes an acme account from storage
func (service *Service) DeleteAccount(w http.ResponseWriter, r *http.Request) *output.Error {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// verify account exists
	_, outErr := service.getAccount(id)
	if outErr != nil {
		return outErr
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

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted acme account (id: %d)", id),
	}

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
