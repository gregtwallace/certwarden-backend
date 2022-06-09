package acme_accounts

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// GetAllAccounts is an http handler that returns all acme accounts in the form of JSON written to w
func (accountsApp *AccountsApp) GetAllAccounts(w http.ResponseWriter, r *http.Request) {

	accounts, err := accountsApp.DB.getAllAccounts()
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetAll: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetAll: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetOneAccount is an http handler that returns one acme account based on its unique id in the
//  form of JSON written to w
func (accountsApp *AccountsApp) GetOneAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// if id is new, provide some info
	err := utils.IsIdValidNew(idParam)
	if err == nil {
		accountsApp.GetNewAccountOptions(w, r)
		return
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetOne: id param issue -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	if id < 0 {
		// if id < 0, it is definitely not valid
		err = errors.New("accounts: GetOne: id param is invalid (less than 0 and not new)")
		accountsApp.Logger.Println(err)
		utils.WriteErrorJSON(w, err)
		return
	}

	acmeAccount, err := accountsApp.DB.getOneAccount(id)
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetOne: db failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

// GetNewAccountOptions is an http handler that returns information the client GUI needs to properly
//  present options when the user is creating an account
func (accountsApp *AccountsApp) GetNewAccountOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish constructing all needed options
	newAccountOptions := NewAccountOptions{}
	newAccountOptions.TosUrl = accountsApp.Acme.ProdDir.Meta.TermsOfService
	newAccountOptions.StagingTosUrl = accountsApp.Acme.StagingDir.Meta.TermsOfService

	availableKeys, err := accountsApp.DB.getAvailableKeys()
	if err != nil {
		accountsApp.Logger.Printf("accounts: GetNewOptions: failed to get available keys -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	newAccountOptions.AvailableKeys = availableKeys

	utils.WriteJSON(w, http.StatusOK, newAccountOptions, "acme_account_options")
}

// PutOneAccount is an http handler that overwrites the specified (by id) acme account with the
//  data PUT by the client

// TODO: fetch data from LE and update kid
// if kid already exists, use it to update the email field

func (accountsApp *AccountsApp) PutOneAccount(w http.ResponseWriter, r *http.Request) {
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")

	var payload accountPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// validation
	// id
	err = utils.IsIdValidExisting(idParam, payload.ID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
	}
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	///

	// check db for account kid
	kid, err := accountsApp.DB.getAccountKid(payload.ID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: failed to check for kid -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// load fields
	account, err := payload.accountPayloadToDb()
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: failed to load payload into db obj -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// if no kid, do new LE registration
	if kid == "" {
		// always update name and desc
		err = accountsApp.DB.putExistingAccountNameDesc(account)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PutOne: failed to update account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// do new LE reg
		err = accountsApp.createNewLEAccount(payload)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PutOne: failed to create new LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	} else {
		// if there is a kid, update the existing account with new data
		// do LE update
		err = accountsApp.updateLEAccount(payload)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PutOne: failed to update LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}

		// update db
		err = accountsApp.DB.putExistingAccount(account)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PutOne: failed to write to db -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	}

	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}

func (accountsApp *AccountsApp) PostNewAccount(w http.ResponseWriter, r *http.Request) {
	var payload accountPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: failed to decode json -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	/// do validation
	// id
	err = utils.IsIdValidNew(payload.ID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: invalid id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// name
	err = utils.IsNameValid(payload.Name)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: invalid name -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	} // check db for duplicate name? probably unneeded as sql will error on insert
	// email
	err = utils.IsEmailValidOrBlank(payload.Email)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: invalid email -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// private key (pem valid is checked later in account creation)
	// TOS must be accepted
	if (payload.AcceptedTos != "true") && (payload.AcceptedTos != "on") {
		accountsApp.Logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	///

	// payload -> LE create -> save to db
	err = accountsApp.createNewAccount(payload)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: failed to create account -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Write OK response to client
	response := utils.JsonResp{
		OK: true,
	}
	err = utils.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: write response json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
