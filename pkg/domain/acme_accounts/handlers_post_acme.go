package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/domain/private_keys/key_crypto"
	"legocerthub-backend/pkg/output"
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
// endpoint: /api/v1/acmeaccounts/:id/new-account
func (service *Service) NewAcmeAccount(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// decode body into payload
	var payload registerPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation (only need to confirm account exists and has a key)
	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// get crypto key
	key, err := key_crypto.PemStringToKey(account.AccountKey.Pem, account.AccountKey.Algorithm)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}
	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	var acmeAccount AcmeAccount
	acmeAccount.Account, err = acmeService.NewAccount(account.newAccountPayload(payload.EabKid, payload.EabHmacKey), key)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// add additional details to the acmeAccount before saving
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
		Message: "registered",
		ID:      idParam,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}

// Deactivate sets deactivated status for the ACME account
// Once deactivated, accounts cannot be re-enabled. This action is DANGEROUS
// and should only be done when there is a complete understanding of the repurcussions.
// endpoint: /api/v1/acmeaccounts/:id/deactivate
func (service *Service) Deactivate(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// validation
	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// get acme AccountKey
	acmeAccountKey, err := account.AcmeAccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// if kid is blank, can't deactivate
	if acmeAccountKey.Kid == "" {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// status must be 'valid' to deactivate
	if account.Status != "valid" {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// end validation

	// send the new-account to ACME
	acmeService, err := service.acmeServerService.AcmeService(account.AcmeServer.ID)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	var acmeAccount AcmeAccount
	acmeAccount.Account, err = acmeService.DeactivateAccount(acmeAccountKey)
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// add additional details to the acmeAccount before saving
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
		Message: "deactivated",
		ID:      idParam,
	}

	err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
