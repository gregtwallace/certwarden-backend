package acme_accounts

import (
	"errors"
	"legocerthub-backend/utils"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

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

func (acmeAccountsApp *AcmeAccountsApp) GetOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetOne: id param issue -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}

	// if id is -1 (new) provide new account options
	if id == -1 {
		// TODO
		return
	} else if id < 0 {
		// if id < 0, it is definitely not valid
		err = errors.New("acmeaccounts: GetOne: id param is invalid (less than 0 and not -1)")
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
