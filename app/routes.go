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
	privateKeys := private_keys.PrivateKeys{
		DB:        app.DB.DB,
		DBTimeout: app.DB.Timeout,
		Logger:    app.Logger,
	}
	router.HandlerFunc(http.MethodGet, "/v1/privatekeys", privateKeys.GetAllPrivateKeys)

	// acme accounts definition and handlers
	acmeAccounts := acme_accounts.AcmeAccounts{
		DB:        app.DB.DB,
		DBTimeout: app.DB.Timeout,
		Logger:    app.Logger,
	}
	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts", acmeAccounts.GetAllAcmeAccounts)
	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts/:id", acmeAccounts.GetOneAcmeAccount)

	return app.enableCORS(router)
}
