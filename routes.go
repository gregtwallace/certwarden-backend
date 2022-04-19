package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/status", app.statusHandler)

	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts", app.getAllAcmeAccounts)
	router.HandlerFunc(http.MethodGet, "/v1/acmeaccount/:id", app.getOneAcmeAccount)

	return app.enableCORS(router)
}
