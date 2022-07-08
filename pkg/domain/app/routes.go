package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) Routes() http.Handler {
	app.router = httprouter.New()

	// app
	app.makeHandle(http.MethodGet, "/api/status", app.statusHandler)

	// private_keys
	app.makeHandle(http.MethodGet, "/api/v1/privatekeys", app.keys.GetAllKeys)
	app.makeHandle(http.MethodPost, "/api/v1/privatekeys", app.keys.PostNewKey)

	app.makeHandle(http.MethodGet, "/api/v1/privatekeys/:id", app.keys.GetOneKey)
	app.makeHandle(http.MethodPut, "/api/v1/privatekeys/:id", app.keys.PutNameDescKey)

	app.makeHandle(http.MethodDelete, "/api/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.router.HandlerFunc(http.MethodPost, "/api/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.router.HandlerFunc(http.MethodGet, "/api/v1/acmeaccounts/:id", app.accounts.GetOneAccount)
	app.router.HandlerFunc(http.MethodPut, "/api/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.router.HandlerFunc(http.MethodPost, "/api/v1/acmeaccounts/:id/new-account", app.accounts.NewAccount)
	app.router.HandlerFunc(http.MethodPost, "/api/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)
	app.router.HandlerFunc(http.MethodPut, "/api/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)

	app.router.HandlerFunc(http.MethodDelete, "/api/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
