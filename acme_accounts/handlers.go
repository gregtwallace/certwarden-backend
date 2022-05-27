package acme_accounts

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// GetAllAcmeAccounts is an http handler that returns all acme accounts in the form of JSON written to w
func (acmeAccountsApp *AcmeAccountsApp) GetAllAcmeAccounts(w http.ResponseWriter, r *http.Request) {

	accounts, err := acmeAccountsApp.dbGetAllAcmeAccounts()
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetOneAcmeAccount is an http handler that returns one acme account based on its unique id in the
//  form of JSON written to w
func (acmeAccountsApp *AcmeAccountsApp) GetOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// if id is new, provide some info
	err := utils.IsIdValidNew(idParam)
	if err == nil {
		acmeAccountsApp.GetNewAccountOptions(w, r)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetOne: id param issue -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	if id < 0 {
		// if id < 0, it is definitely not valid
		err = errors.New("acmeaccounts: GetOne: id param is invalid (less than 0 and not new)")
		acmeAccountsApp.Logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	acmeAccount, err := acmeAccountsApp.dbGetOneAcmeAccount(id)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
//  present options when the user is creating an account
func (acmeAccountsApp *AcmeAccountsApp) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish constructing all needed options
	newAccountOptions := NewAcmeAccountOptions{}
	newAccountOptions.TosUrl = acmeAccountsApp.Acme.ProdDir.Meta.TermsOfService
	newAccountOptions.StagingTosUrl = acmeAccountsApp.Acme.StagingDir.Meta.TermsOfService

	availableKeys, err := acmeAccountsApp.dbGetAvailableKeys()
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetNewOptions: failed to get available keys -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	newAccountOptions.AvailableKeys = availableKeys

	utils.WriteJSON(w, http.StatusOK, newAccountOptions, "acme_account_options")
}

// PutOneAcmeAccount is an http handler that overwrites the specified (by id) acme account with the
//  data PUT by the client
func (acmeAccountsApp *AcmeAccountsApp) PutOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	var payload acmeAccountPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = utils.IsIdValidExisting(idParam, payload.ID)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
	}
	///

	// load fields
	var acmeAccount acmeAccountDb
	acmeAccount.id, err = strconv.Atoi(payload.ID)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	acmeAccount.name = payload.Name

	acmeAccount.description.Valid = true
	acmeAccount.description.String = payload.Description

	acmeAccount.email.Valid = true
	acmeAccount.email.String = payload.Email

	acmeAccount.updatedAt = int(time.Now().Unix())

	// once tos is accepted, it cannot be un-accepted
	// logic to prevent un-accepting
	if payload.AcceptedTos == "on" || payload.AcceptedTos == "true" {
		acmeAccount.acceptedTos.Bool = true
		acmeAccount.acceptedTos.Valid = true
	} else {
		acmeAccount.acceptedTos.Valid = false
	}

	err = acmeAccountsApp.dbPutExistingAcmeAccount(acmeAccount)
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

}
