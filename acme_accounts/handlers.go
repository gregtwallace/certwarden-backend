package acme_accounts

import (
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

	acmeAccount := acmeAccount{
		ID:           id,
		PrivateKeyID: 1,
		Name:         "Another Acct",
		Email:        "something@test.com",
		Description:  "Staging 1",
		IsStaging:    true,
	}

	err = utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
	if err != nil {
		acmeAccountsApp.Logger.Printf("acmeaccounts: GetOne: write json failed -- err: %s", err)
		utils.WriteErrorJSON(w, err)
		return
	}
}
