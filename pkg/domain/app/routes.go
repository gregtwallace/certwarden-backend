package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) Routes() http.Handler {
	app.router = httprouter.New()

	// app - insecure
	app.makeHandle(http.MethodPost, "/api/v1/auth/login", app.auth.Login)
	app.makeHandle(http.MethodPost, "/api/v1/auth/refresh", app.auth.Refresh)

	// app
	app.makeSecureHandle(http.MethodGet, "/api/status", app.statusHandler)

	// private_keys
	app.makeSecureHandle(http.MethodGet, "/api/v1/privatekeys", app.keys.GetAllKeys)
	app.makeSecureHandle(http.MethodGet, "/api/v1/privatekeys/:id", app.keys.GetOneKey)

	app.makeSecureHandle(http.MethodPost, "/api/v1/privatekeys", app.keys.PostNewKey)

	app.makeSecureHandle(http.MethodPut, "/api/v1/privatekeys/:id", app.keys.PutNameDescKey)

	app.makeSecureHandle(http.MethodDelete, "/api/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.makeSecureHandle(http.MethodGet, "/api/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.makeSecureHandle(http.MethodGet, "/api/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	app.makeSecureHandle(http.MethodPost, "/api/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.makeSecureHandle(http.MethodPut, "/api/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.makeSecureHandle(http.MethodPut, "/api/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)

	app.makeSecureHandle(http.MethodPost, "/api/v1/acmeaccounts/:id/new-account", app.accounts.NewAcmeAccount)
	app.makeSecureHandle(http.MethodPost, "/api/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)

	app.makeSecureHandle(http.MethodDelete, "/api/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	app.makeSecureHandle(http.MethodGet, "/api/v1/certificates", app.certificates.GetAllCerts)
	app.makeSecureHandle(http.MethodGet, "/api/v1/certificates/:certid", app.certificates.GetOneCert)

	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates", app.certificates.PostNewCert)

	app.makeSecureHandle(http.MethodPut, "/api/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.makeSecureHandle(http.MethodDelete, "/api/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.makeSecureHandle(http.MethodGet, "/api/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders", app.orders.NewOrder)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.makeHandle(http.MethodGet, "/api/v1/download/privatekeys/:name", app.keys.GetKeyPemFile)
	app.makeHandle(http.MethodGet, "/api/v1/download/certificates/:name", app.certificates.GetCertPemFile)

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
