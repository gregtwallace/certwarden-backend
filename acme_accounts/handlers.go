package acme_accounts

import (
	"errors"
	"legocerthub-backend/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (acmeAccounts *AcmeAccounts) GetAllAcmeAccounts(w http.ResponseWriter, r *http.Request) {

	accounts, err := acmeAccounts.dbGetAllAcmeAccounts()
	if err != nil {
		log.Printf("Failed to get all ACME accounts %s", err)
	}

	utils.WriteJSON(w, http.StatusOK, accounts, "acme_accounts")

}

func (acmeAccounts *AcmeAccounts) GetOneAcmeAccount(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		log.Print(errors.New("invalid id parameter"))
		//utils.errorJSON(w, err)
		return
	}

	acmeAccount := acmeAccount{
		ID:           id,
		PrivateKeyID: 10,
		Name:         "Another Acct",
		Email:        "something@test.com",
		Description:  "Staging 1",
		IsStaging:    true,
	}

	utils.WriteJSON(w, http.StatusOK, acmeAccount, "acme_account")
}
