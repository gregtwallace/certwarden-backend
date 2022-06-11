package app

import (
	"legocerthub-backend/pkg/acme_accounts"
	"legocerthub-backend/pkg/private_keys"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	// app handlers (app already defined)
	router.HandlerFunc(http.MethodGet, "/api/status", app.statusHandler)

	// private keys service
	keyService := private_keys.NewService(app)

	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys", keyService.GetAllKeys)
	router.HandlerFunc(http.MethodPost, "/api/v1/privatekeys", keyService.PostNewKey)
	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys/:id", keyService.GetOneKey)
	router.HandlerFunc(http.MethodPut, "/api/v1/privatekeys/:id", keyService.PutOneKey)
	router.HandlerFunc(http.MethodDelete, "/api/v1/privatekeys/:id", keyService.DeleteKey)

	// TODO MODIFY apps to have receiver for the stuff we pass in, as opposed to specific types
	// e.g. func PrivateKeysApp.New which accepts an Interface as long as it implements the needed pieces
	//   can make funcs such as .Logger() that returns the logger (assuming forced to use methods instead of types)

	// acme accounts definition and handlers
	acmeAccounts := acme_accounts.AccountsApp{}
	acmeAccounts.Logger = app.Logger
	acmeAccounts.DB.Database = app.Storage.Db
	acmeAccounts.DB.Timeout = app.Storage.Timeout
	acmeAccounts.Acme.ProdDir = app.Acme.ProdDir
	acmeAccounts.Acme.StagingDir = app.Acme.StagingDir

	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts", acmeAccounts.GetAllAccounts)
	router.HandlerFunc(http.MethodPost, "/api/v1/acmeaccounts", acmeAccounts.PostNewAccount)
	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts/:id", acmeAccounts.GetOneAccount)
	router.HandlerFunc(http.MethodPut, "/api/v1/acmeaccounts/:id", acmeAccounts.PutOneAccount)

	return app.enableCORS(router)
}
