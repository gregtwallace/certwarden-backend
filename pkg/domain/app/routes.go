package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) routes() http.Handler {
	app.router = httprouter.New()

	// auth - insecure
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/login", app.auth.Login)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/refresh", app.auth.Refresh)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/logout", app.auth.Logout)
	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/auth/changepassword", app.auth.ChangePassword)

	// health check (HEAD or GET) - insecure
	app.makeHandle(http.MethodHead, apiUrlPath+"/health", app.healthHandler)
	app.makeHandle(http.MethodGet, apiUrlPath+"/health", app.healthHandler)

	// app
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/status", app.statusHandler)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/log", app.viewCurrentLogHandler)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/logs", app.downloadLogsHandler)

	// updater
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/app/new-version", app.updater.GetNewVersionInfo)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/app/new-version", app.updater.CheckForNewVersion)

	// acme_servers
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeservers", app.acmeServers.GetAllServers)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.GetOneServer)

	// private_keys
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys", app.keys.GetAllKeys)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id", app.keys.GetOneKey)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id/download", app.keys.DownloadOneKey)

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/privatekeys", app.keys.PostNewKey)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.StageNewApiKey)
	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.RemoveOldApiKey)

	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/privatekeys/:id", app.keys.PutKeyUpdate)

	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)
	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/key-change", app.accounts.RolloverKey)

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/new-account", app.accounts.NewAcmeAccount)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)

	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates", app.certificates.GetAllCerts)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid", app.certificates.GetOneCert)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/download", app.certificates.DownloadOneCert)

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates", app.certificates.PostNewCert)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.StageNewApiKey)
	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.RemoveOldApiKey)

	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.NewOrder)

	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/download", app.orders.DownloadOneOrder)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatekeys/:name", app.download.DownloadKeyViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certificates/:name", app.download.DownloadCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certrootchains/:name", app.download.DownloadCertRootChainViaHeader)

	// download keys and certs - via URL routes
	// include
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatekeys/:name/*apiKey", app.download.DownloadKeyViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certificates/:name/*apiKey", app.download.DownloadCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatecerts/:name/*apiKey", app.download.DownloadPrivateCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certrootchains/:name/*apiKey", app.download.DownloadCertRootChainViaUrl)

	// frontend (if enabled)
	if *app.config.ServeFrontend {
		// log availability
		app.logger.Infof("frontend hosting enabled and available at: %s", frontendUrlPath)

		// configure environment file
		app.setFrontendEnv()

		// redirect root to frontend app
		app.makeHandle(http.MethodGet, "/", redirectToFrontendHandler)

		// add file server route for frontend
		app.makeHandle(http.MethodGet, frontendUrlPath+"/*anything", app.frontendHandler)
	}

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)
	app.router.MethodNotAllowed = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
