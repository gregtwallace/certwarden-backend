package acme_accounts

import (
	"encoding/json"
	"errors"
	"legocerthub-backend/utils"
	"legocerthub-backend/utils/acme_utils"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	///

	// load fields
	var acmeAccount accountDb
	acmeAccount.id, err = strconv.Atoi(payload.ID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: invalid id -- err: %s", err)
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

	err = accountsApp.DB.putExistingAccount(acmeAccount)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PutOne: failed to write to db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
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
	// private key - get pem (only verifies private key id exists)
	keyPem, err := accountsApp.DB.getKeyPem(payload.PrivateKeyID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: invalid private key id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	// TOS must be accepted
	if (payload.AcceptedTos != "true") && (payload.AcceptedTos != "on") {
		accountsApp.Logger.Println("accounts: PostNew: must accept ToS")
		utils.WriteErrorJSON(w, errors.New("accounts: PostNew: must accept ToS"))
		return
	}
	///

	// load fields
	var account accountDb
	account.name = payload.Name

	account.description.Valid = true
	account.description.String = payload.Description

	keyId, err := strconv.Atoi(payload.PrivateKeyID)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: invalid key id -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
	account.privateKeyId.Valid = true
	account.privateKeyId.Int32 = int32(keyId)

	account.status.Valid = true
	account.status.String = "Unknown"

	account.email.Valid = true
	account.email.String = payload.Email

	account.acceptedTos.Valid = true
	account.acceptedTos.Bool = true

	account.isStaging.Valid = true
	if (payload.IsStaging == "true") || (payload.IsStaging == "on") {
		account.isStaging.Bool = true
	} else {
		account.isStaging.Bool = false
	}

	account.createdAt = 0 // will update later from LE
	account.updatedAt = int(time.Now().Unix())

	err = accountsApp.DB.postNewAccount(account)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: failed to write db -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// Create account with LE
	var acmeAccount acme_utils.AcmeAccount
	acmeAccount.Contact = []string{"mailto:" + payload.Email}
	acmeAccount.TermsOfServiceAgreed = true

	var acmeAccountResponse acme_utils.AcmeAccountResponse
	if (payload.IsStaging == "true") || (payload.IsStaging == "on") {
		acmeAccountResponse, err = accountsApp.Acme.StagingDir.CreateAccount(acmeAccount, keyPem)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PostNew: failed to create LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	} else {
		acmeAccountResponse, err = accountsApp.Acme.ProdDir.CreateAccount(acmeAccount, keyPem)
		if err != nil {
			accountsApp.Logger.Printf("accounts: PostNew: failed to create LE account -- err: %s", err)
			utils.WriteErrorJSON(w, err)
			return
		}
	}

	// Write the returned account info from LE to the db
	account.email.String = strings.TrimPrefix(acmeAccountResponse.Contact[0], "mailto:")

	unixCreated, err := acme_utils.LeToUnixTime(acmeAccountResponse.CreatedAt)
	if err != nil {
		unixCreated = 0
	}
	account.createdAt = int(unixCreated)

	account.updatedAt = int(time.Now().Unix())

	// already valid
	account.status.String = acmeAccountResponse.Status

	account.kid.Valid = true
	account.kid.String = acmeAccountResponse.Location

	// update the account with LE info
	err = accountsApp.DB.putLEAccountInfo(account)
	if err != nil {
		accountsApp.Logger.Printf("accounts: PostNew: failed to update db -- err: %s", err)
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
