package acme_accounts

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// DeleteAccount deletes an acme account from storage
func (service *Service) DeleteAccount(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// verify account exists
	_, outErr := service.getAccount(id)
	if outErr != nil {
		return outErr
	}

	// do not allow delete if in use
	inUse, err := service.storage.AcmeAccountInUse(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}
	if inUse {
		service.logger.Debug("cannot delete, in use")
		return output.ErrorJsonErrDeleteInUse("acme account")
	}
	// end validation

	// delete from storage
	err = service.storage.DeleteAcmeAccount(id)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	// write response
	response := &output.JsonResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("deleted acme account (id: %d)", id),
	}

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
