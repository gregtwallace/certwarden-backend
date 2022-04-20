package application

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/status", app.StatusHandler)

	router.HandlerFunc(http.MethodGet, "/v1/acmeaccounts", app.GetAllAcmeAccounts)
	router.HandlerFunc(http.MethodGet, "/v1/acmeaccount/:id", app.GetOneAcmeAccount)

	return app.enableCORS(router)
}
