package app

import (
	"errors"
	"legocerthub-backend/database"
	"legocerthub-backend/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) getAllAcmeAccounts(w http.ResponseWriter, r *http.Request) {

	accounts, err := app.Database.DBGetAllAcmeAccounts()
	if err != nil {
		log.Printf("Failed to get all ACME accounts %s", err)
	}

	utils.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")

}

func (app *Application) getOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		log.Print(errors.New("invalid id parameter"))
		//utils.errorJSON(w, err)
		return
	}

	acmeAccount := database.AcmeAccount{
		ID:           id,
		PrivateKeyID: 10,
		Name:         "Another Acct",
		Email:        "something@test.com",
		Description:  "Staging 1",
		IsStaging:    true,
	}

	utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
}
