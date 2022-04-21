package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/status", app.statusHandler)

	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts", app.getAllAcmeAccounts)
	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts/:id", app.getOneAcmeAccount)

	return app.enableCORS(router)
}
