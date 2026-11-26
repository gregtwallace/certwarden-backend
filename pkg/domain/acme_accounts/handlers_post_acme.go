package acme_accounts

import (
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/validation"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// register payload contains External Account Binding information (if required)
type registerPayload struct {
	Email      *string `json:"email"`
	EabKid     string  `json:"eab_kid"`
	EabHmacKey string  `json:"eab_hmac_key"`
}

// NewAcmeAccount sends the account information to the ACME new-account endpoint
// which effectively registers the account with ACME
func (service *Service) NewAcmeAccount(w http.ResponseWriter, r *http.Request) *output.JsonError {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// decode body into payload
	var payload registerPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation (only need to confirm account exists and has a key)
	// fetch the relevant account
	account, err := service.storage.GetOneAcmeAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// get crypto key
	key, err := key_crypto.PemStringToKey(account.AccountKey.Pem, account.AccountKey.Algorithm)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// validate override email if it exists
	if payload.Email != nil && *payload.Email != "" && !validation.EmailValid(*payload.Email) {
		service.logger.Debug(ErrEmailBad)
		return output.ErrorJsonErrValidationFailed(ErrEmailBad)
	}

	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	acmeAcct, err := acmeService.NewAccount(account.newAccountPayload(payload.EabKid, payload.EabHmacKey, payload.Email), key)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// save ACME response to storage
	account, err = service.storage.PutAcmeAccountUpdate(acmeAcctToUpdatePayload(idParam, acmeAcct))
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := account.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "registered account"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}

// RefreshAcmeAccount gets the current state of the ACME Account object from the
// ACME Server and updates it in the database. The object is also returned to the
// client.
func (service *Service) RefreshAcmeAccount(w http.ResponseWriter, r *http.Request) *output.JsonError {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation - confirm account exists and has a kid / URL
	// fetch the relevant account
	account, err := service.storage.GetOneAcmeAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// get acme AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// if kid/url is blank, can't GET account object
	if acmeAccountKey.Kid == "" {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}
	// end validation

	// GET from ACME server
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	acmeAcct, err := acmeService.GetAccount(acmeAccountKey)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// save ACME response to storage
	account, err = service.storage.PutAcmeAccountUpdate(acmeAcctToUpdatePayload(idParam, acmeAcct))
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := account.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "account state fetched and updated"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}

// Deactivate sets deactivated status for the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
// endpoint: /api/v1/acmeaccounts/:id/deactivate
func (service *Service) Deactivate(w http.ResponseWriter, r *http.Request) *output.JsonError {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// validation
	// fetch the relevant account
	account, err := service.storage.GetOneAcmeAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	// get acme AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// if kid is blank, can't deactivate
	if acmeAccountKey.Kid == "" {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}

	// status must be 'valid' to deactivate
	if account.Status != "valid" {
		service.logger.Debug(err)
		return output.ErrorJsonErrValidationFailed(err)
	}
	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	acmeAcct, err := acmeService.DeactivateAccount(acmeAccountKey)
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// save ACME response to storage
	account, err = service.storage.PutAcmeAccountUpdate(acmeAcctToUpdatePayload(idParam, acmeAcct))
	if err != nil {
		service.logger.Error(err)
		return output.ErrorJsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := account.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%w)", err)
		service.logger.Error(err)
		return output.ErrorJsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "deactivated account"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}
