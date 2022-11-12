package acme_accounts

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllAccounts is an http handler that returns all acme accounts in the form of JSON written to w
func (service *Service) GetAllAccounts(w http.ResponseWriter, r *http.Request) (err error) {
	// get all from storage
	accounts, err := service.storage.GetAllAccounts()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// make keysResponse (for json output)
	var acctsResponse []AccountSummaryResponse
	for i := range accounts {
		acctsResponse = append(acctsResponse, accounts[i].SummaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, acctsResponse, "acme_accounts")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetOneAccount is an http handler that returns one acme account based on its unique id in the
// form of JSON written to w
func (service *Service) GetOneAccount(w http.ResponseWriter, r *http.Request) (err error) {
	// get id from param
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	if validation.IsIdNew(id) {
		return service.GetNewAccountOptions(w, r)
	}

	// get from storage
	account, err := service.getAccount(id)
	if err != nil {
		return err
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, account.detailedResponse(), "acme_account")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
// present options when the user is creating an account
func (service *Service) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) (err error) {
	// account options / info to assist client with new account posting
	newAccountOptions := newAccountOptions{}

	// tos
	newAccountOptions.TosUrl = service.acmeProd.TosUrl()
	newAccountOptions.StagingTosUrl = service.acmeStaging.TosUrl()

	// available private keys
	keys, err := service.keys.AvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	for i := range keys {
		newAccountOptions.AvailableKeys = append(newAccountOptions.AvailableKeys, keys[i].SummaryResponse())
	}

	// return response to client
	_, err = service.output.WriteJSON(w, http.StatusOK, newAccountOptions, "acme_account_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
