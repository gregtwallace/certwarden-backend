package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NewAcmeAccount sends the account information to the ACME new-account endpoint
// which effectively registers the account with ACME
// endpoint: /api/v1/acmeaccounts/:id/new-account
func (service *Service) NewAcmeAccount(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam, true)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// no need to validate, can try to register any account
	// that has a cryptokey

	// get crypto key
	key, err := account.PrivateKey.CryptoKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if *account.IsStaging {
		acmeResponse, err = service.acmeStaging.NewAccount(account.newAccountPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewAccount(account.newAccountPayload(), key)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
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

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
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

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam, true)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// no need to validate, can try to deactivate any account
	// that has an accountKey

	// get AccountKey
	accountKey, err := account.AccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// send the new-account to ACME
	var acmeResponse acme.AcmeAccountResponse
	if *account.IsStaging {
		acmeResponse, err = service.acmeStaging.DeactivateAccount(accountKey)
	} else {
		acmeResponse, err = service.acmeProd.DeactivateAccount(accountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
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

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
