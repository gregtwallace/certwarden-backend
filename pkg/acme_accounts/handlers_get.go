package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllAccounts is an http handler that returns all acme accounts in the form of JSON written to w
func (service *Service) GetAllAccounts(w http.ResponseWriter, r *http.Request) {

	accounts, err := service.storage.GetAllAccounts()
	if err != nil {
		service.logger.Printf("accounts: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")
	if err != nil {
		service.logger.Printf("accounts: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetOneAccount is an http handler that returns one acme account based on its unique id in the
//  form of JSON written to w
func (service *Service) GetOneAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// if id is new, provide some info
	err := utils.IsIdValidNew(idParam)
	if err == nil {
		service.GetNewAccountOptions(w, r)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		service.logger.Printf("accounts: GetOne: id param issue -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	if id < 0 {
		// if id < 0, it is definitely not valid
		err = errors.New("accounts: GetOne: id param is invalid (less than 0 and not new)")
		service.logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	acmeAccount, err := service.storage.GetOneAccount(id)
	if err != nil {
		service.logger.Printf("accounts: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
	if err != nil {
		service.logger.Printf("accounts: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
//  present options when the user is creating an account
func (service *Service) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish constructing all needed options
	newAccountOptions := newAccountOptions{}
	newAccountOptions.TosUrl = service.acmeProdDir.Meta.TermsOfService
	newAccountOptions.StagingTosUrl = service.acmeStagingDir.Meta.TermsOfService

	availableKeys, err := service.storage.GetAvailableKeys()
	if err != nil {
		service.logger.Printf("accounts: GetNewOptions: failed to get available keys -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	newAccountOptions.AvailableKeys = availableKeys

	utils.WriteJSON(w, http.StatusOK, newAccountOptions, "acme_account_options")
}
