package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// NameDescPayload is the struct for editing an existing account name and desc
type NameDescPayload struct {
	ID          int     `json:"-"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	UpdatedAt   int     `json:"-"`
}

// PutNameDescAccount is a handler that sets the name and description of an account
// within storage
func (service *Service) PutNameDescAccount(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// payload decoding
	var payload NameDescPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	if !service.idValid(idParam) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// name (optional)
	if payload.Name != nil && !service.nameValid(*payload.Name, &idParam) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// end validation

	// add additional details to the payload before saving
	payload.ID = idParam
	payload.UpdatedAt = int(time.Now().Unix())

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	err = service.storage.PutNameDescAccount(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      idParam,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// ChangeEmailPayload is the struct for updating an account's email address
type ChangeEmailPayload struct {
	Email *string `json:"email"`
}

// ChangeEmail() is a handler that updates an ACME account with the specified
// email address and saves the updated address to storage
func (service *Service) ChangeEmail(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// decode payload
	var payload ChangeEmailPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// id
	if !service.idValid(idParam) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// email (update cannot be to blank)
	if payload.Email == nil || !validation.EmailValid(*payload.Email) {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// end validation

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// get AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// make ACME update email payload
	acmePayload := acme.UpdateAccountPayload{
		Contact: emailToContact(*payload.Email),
	}

	// send the email update to ACME
	var acmeAccount AcmeAccount
	if account.IsStaging {
		acmeAccount.Account, err = service.acmeStaging.UpdateAccount(acmePayload, acmeAccountKey)
	} else {
		acmeAccount.Account, err = service.acmeProd.UpdateAccount(acmePayload, acmeAccountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// add additional details to the payload before saving
	acmeAccount.ID = idParam
	acmeAccount.UpdatedAt = int(time.Now().Unix())

	// save ACME response to account
	err = service.storage.PutAcmeAccountResponse(acmeAccount)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      idParam,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
