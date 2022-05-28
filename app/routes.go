package app

import (
	"legocerthub-backend/acme_accounts"
	"legocerthub-backend/private_keys"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	// app handlers (app already defined)
	router.HandlerFunc(http.MethodGet, "/api/status", app.statusHandler)

	// private keys definition and handlers
	privateKeys := private_keys.KeysApp{}
	privateKeys.Logger = app.Logger
	privateKeys.DB.Database = app.DB.Database
	privateKeys.DB.Timeout = app.DB.Timeout

	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys/", privateKeys.GetAllKeys)
	router.HandlerFunc(http.MethodPost, "/api/v1/privatekeys/", privateKeys.PostNewKey)
	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys/:id", privateKeys.GetOneKey)
	router.HandlerFunc(http.MethodPut, "/api/v1/privatekeys/:id", privateKeys.PutOneKey)
	router.HandlerFunc(http.MethodDelete, "/api/v1/privatekeys/:id", privateKeys.DeleteKey)

	// acme accounts definition and handlers
	acmeAccounts := acme_accounts.AccountsApp{}
	acmeAccounts.Logger = app.Logger
	acmeAccounts.DB.Database = app.DB.Database
	acmeAccounts.DB.Timeout = app.DB.Timeout
	acmeAccounts.Acme.ProdDir = app.Acme.ProdDir
	acmeAccounts.Acme.StagingDir = app.Acme.StagingDir

	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts", acmeAccounts.GetAllAccounts)
	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts/:id", acmeAccounts.GetOneAccount)
	router.HandlerFunc(http.MethodPut, "/api/v1/acmeaccounts/:id", acmeAccounts.PutOneAccount)

	return app.enableCORS(router)
}
