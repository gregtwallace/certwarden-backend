package acme_accounts

import (
	"certwarden-backend/pkg/domain/private_keys/key_crypto"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// register payload contains External Account Binding information (if required)
type registerPayload struct {
	EabKid     string `json:"eab_kid"`
	EabHmacKey string `json:"eab_hmac_key"`
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
		return output.JsonErrValidationFailed(err)
	}

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// validation (only need to confirm account exists and has a key)
	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrValidationFailed(err)
	}

	// get crypto key
	key, err := key_crypto.PemStringToKey(account.AccountKey.Pem, account.AccountKey.Algorithm)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}
	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	var acmeAccount AcmeAccount
	acmeAccount.Account, err = acmeService.NewAccount(account.newAccountPayload(payload.EabKid, payload.EabHmacKey), key)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// add additional details to the acmeAccount before saving
	acmeAccount.ID = idParam
	acmeAccount.UpdatedAt = int(time.Now().Unix())

	// save ACME response to account
	updatedAcct, err := service.storage.PutAcmeAccountResponse(acmeAccount)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := updatedAcct.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "registered account"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
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
		return output.JsonErrValidationFailed(err)
	}

	// validation - confirm account exists and has a kid / URL
	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrValidationFailed(err)
	}

	// get acme AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// if kid/url is blank, can't GET account object
	if acmeAccountKey.Kid == "" {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	// end validation

	// GET from ACME server
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	var acmeAccount AcmeAccount
	acmeAccount.Account, err = acmeService.GetAccount(acmeAccountKey)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// add additional details to the acmeAccount before saving
	acmeAccount.ID = idParam
	acmeAccount.UpdatedAt = int(time.Now().Unix())

	// save ACME response to account
	updatedAcct, err := service.storage.PutAcmeAccountResponse(acmeAccount)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := updatedAcct.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "account state fetched and updated"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
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
		return output.JsonErrValidationFailed(err)
	}

	// validation
	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	// get acme AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// if kid is blank, can't deactivate
	if acmeAccountKey.Kid == "" {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// status must be 'valid' to deactivate
	if account.Status != "valid" {
		service.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}
	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	var acmeAccount AcmeAccount
	acmeAccount.Account, err = acmeService.DeactivateAccount(acmeAccountKey)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// add additional details to the acmeAccount before saving
	acmeAccount.ID = idParam
	acmeAccount.UpdatedAt = int(time.Now().Unix())

	// save ACME response to account
	updatedAcct, err := service.storage.PutAcmeAccountResponse(acmeAccount)
	if err != nil {
		service.logger.Error(err)
		return output.JsonErrStorageGeneric(err)
	}

	updatedAcctDetailedResp, err := updatedAcct.detailedResponse(service)
	if err != nil {
		err = fmt.Errorf("failed to generate account summary response (%s)", err)
		service.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &accountResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "deactivated account"
	response.Account = updatedAcctDetailedResp

	err = service.output.WriteJSON(w, response)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
