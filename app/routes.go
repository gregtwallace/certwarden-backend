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
	router.HandlerFunc(http.MethodGet, "/status", app.statusHandler)

	// private keys definition and handlers
	privateKeys := private_keys.PrivateKeysApp{
		Database: app.DB.Database,
		Timeout:  app.DB.Timeout,
		Logger:   app.Logger,
	}
	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys/", privateKeys.GetAllPrivateKeys)
	router.HandlerFunc(http.MethodPost, "/api/v1/privatekeys/", privateKeys.PostNewPrivateKey)
	router.HandlerFunc(http.MethodGet, "/api/v1/privatekeys/:id", privateKeys.GetOnePrivateKey)
	router.HandlerFunc(http.MethodPut, "/api/v1/privatekeys/:id", privateKeys.PutOnePrivateKey)
	router.HandlerFunc(http.MethodDelete, "/api/v1/privatekeys/:id", privateKeys.DeletePrivateKey)

	// acme accounts definition and handlers
	acmeAccounts := acme_accounts.AcmeAccountsApp{
		Database: app.DB.Database,
		Timeout:  app.DB.Timeout,
		Logger:   app.Logger,
	}
	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts", acmeAccounts.GetAllAcmeAccounts)
	router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts/:id", acmeAccounts.GetOneAcmeAccount)

	return app.enableCORS(router)
}
