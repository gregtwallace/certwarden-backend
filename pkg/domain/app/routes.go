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
	app.makeHandle(http.MethodGet, "/api/v1/privatekeys/:id", app.keys.GetOneKey)

	app.makeHandle(http.MethodPost, "/api/v1/privatekeys", app.keys.PostNewKey)

	app.makeHandle(http.MethodPut, "/api/v1/privatekeys/:id", app.keys.PutNameDescKey)

	app.makeHandle(http.MethodDelete, "/api/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.makeHandle(http.MethodGet, "/api/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.makeHandle(http.MethodGet, "/api/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	app.makeHandle(http.MethodPost, "/api/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.makeHandle(http.MethodPut, "/api/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.makeHandle(http.MethodPut, "/api/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)

	app.makeHandle(http.MethodPost, "/api/v1/acmeaccounts/:id/new-account", app.accounts.NewAcmeAccount)
	app.makeHandle(http.MethodPost, "/api/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)

	app.makeHandle(http.MethodDelete, "/api/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	app.makeHandle(http.MethodGet, "/api/v1/certificates", app.certificates.GetAllCerts)
	app.makeHandle(http.MethodGet, "/api/v1/certificates/:certid", app.certificates.GetOneCert)

	app.makeHandle(http.MethodPost, "/api/v1/certificates", app.certificates.PostNewCert)

	app.makeHandle(http.MethodPut, "/api/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.makeHandle(http.MethodDelete, "/api/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.makeHandle(http.MethodGet, "/api/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeHandle(http.MethodPost, "/api/v1/certificates/:certid/orders", app.orders.NewOrder)
	app.makeHandle(http.MethodPost, "/api/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
