package acme_accounts

import (
	"database/sql"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/utils"
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

	// return response to client
	_, err = output.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetOneAccount is an http handler that returns one acme account based on its unique id in the
//  form of JSON written to w
func (service *Service) GetOneAccount(w http.ResponseWriter, r *http.Request) (err error) {
	// convert id param to an integer
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// if id is new, provide some info
	err = utils.IsIdNew(&id)
	if err == nil {
		return service.GetNewAccountOptions(w, r)
	}

	// if id < 0 & not new, it is definitely not valid
	err = utils.IsIdExisting(&id)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// get from storage
	account, err := service.storage.GetOneAccountById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			service.logger.Debug(err)
			return output.ErrNotFound
		} else {
			service.logger.Error(err)
			return output.ErrStorageGeneric
		}
	}

	// return response to client
	_, err = output.WriteJSON(w, http.StatusOK, account, "acme_account")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
//  present options when the user is creating an account
func (service *Service) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) (err error) {
	// account options / info to assist client with new account posting
	newAccountOptions := newAccountOptions{}

	// tos
	newAccountOptions.TosUrl = service.acmeProd.TosUrl()
	newAccountOptions.StagingTosUrl = service.acmeStaging.TosUrl()

	// available private keys
	newAccountOptions.AvailableKeys, err = service.keys.GetAvailableKeys()
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	_, err = output.WriteJSON(w, http.StatusOK, newAccountOptions, "acme_account_options")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
