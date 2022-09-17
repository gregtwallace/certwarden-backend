package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) routes() http.Handler {
	app.router = httprouter.New()

	// app - auth - insecure
	app.makeHandle(http.MethodPost, "/api/v1/auth/login", app.auth.Login)
	app.makeHandle(http.MethodPost, "/api/v1/auth/refresh", app.auth.Refresh)
	app.makeHandle(http.MethodPost, "/api/v1/auth/logout", app.auth.Logout)
	app.makeHandle(http.MethodPut, "/api/v1/auth/changepassword", app.auth.ChangePassword)

	// app
	app.makeSecureHandle(http.MethodGet, "/api/status", app.statusHandler)

	// private_keys
	app.makeSecureHandle(http.MethodGet, "/api/v1/privatekeys", app.keys.GetAllKeys)
	app.makeSecureHandle(http.MethodGet, "/api/v1/privatekeys/:id", app.keys.GetOneKey)

	app.makeSecureHandle(http.MethodPost, "/api/v1/privatekeys", app.keys.PostNewKey)

	app.makeSecureHandle(http.MethodPut, "/api/v1/privatekeys/:id", app.keys.PutInfoKey)

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
	app.makeSecureHandle(http.MethodGet, "/api/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	app.makeSecureHandle(http.MethodGet, "/api/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders", app.orders.NewOrder)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.makeSecureHandle(http.MethodPost, "/api/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.makeHandle(http.MethodGet, "/api/v1/download/privatekeys/:name", app.download.DownloadKeyViaHeader)
	app.makeHandle(http.MethodGet, "/api/v1/download/certificates/:name", app.download.DownloadCertViaHeader)
	app.makeHandle(http.MethodGet, "/api/v1/download/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	app.makeHandle(http.MethodGet, "/api/v1/download/certificaterootchains/:name", app.download.DownloadCertRootChainViaHeader)

	// download keys and certs - via URL routes
	app.makeHandle(http.MethodGet, "/api/v1/download/privatekeys/:name/:apiKey", app.download.DownloadKeyViaUrl)
	app.makeHandle(http.MethodGet, "/api/v1/download/certificates/:name/:apiKey", app.download.DownloadCertViaUrl)
	app.makeHandle(http.MethodGet, "/api/v1/download/privatecerts/:name/:apiKey", app.download.DownloadPrivateCertViaUrl)
	app.makeHandle(http.MethodGet, "/api/v1/download/certificaterootchains/:name/:apiKey", app.download.DownloadCertRootChainViaUrl)

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
