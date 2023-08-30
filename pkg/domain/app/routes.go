package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// base path
const baseUrlPath = "/legocerthub"

// backend api paths
const apiUrlPath = baseUrlPath + "/api"
const apiDownloadUrlPath = apiUrlPath + "/v1/download"

// frontend React app path (e.g. Vite config `base`)
const frontendUrlPath = baseUrlPath + "/app"

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) routes() http.Handler {
	app.router = httprouter.New()

	// app auth - insecure
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/app/auth/login", app.auth.Login)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/app/auth/refresh", app.auth.Refresh)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/app/auth/logout", app.auth.Logout)
	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/app/auth/changepassword", app.auth.ChangePassword)

	// status & health check (HEAD or GET) - insecure
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/status", app.statusHandler)
	app.makeHandle(http.MethodHead, apiUrlPath+"/health", app.healthHandler)
	app.makeHandle(http.MethodGet, apiUrlPath+"/health", app.healthHandler)

	// app
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/app/log", app.viewCurrentLogHandler)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/app/logs", app.downloadLogsHandler)

	// app control
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/app/control/shutdown", app.doShutdownHandler)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/app/control/restart", app.doRestartHandler)

	// app updater
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/app/updater/new-version", app.updater.GetNewVersionInfo)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/app/updater/new-version", app.updater.CheckForNewVersion)

	// challenges (config)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers", app.challenges.GetProvidersConfigs)

	// acme_servers
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeservers", app.acmeServers.GetAllServers)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.GetOneServer)

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/acmeservers", app.acmeServers.PostNewServer)
	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.PutServerUpdate)
	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.DeleteServer)

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

	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates", app.certificates.PostNewCert)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.StageNewApiKey)
	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.RemoveOldApiKey)

	app.makeSecureHandle(http.MethodPut, apiUrlPath+"/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.makeSecureHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.NewOrder)

	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/download", app.orders.DownloadCertNewestOrder)
	app.makeSecureHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/download", app.orders.DownloadOneOrder)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.makeSecureHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/privatekeys/:name", app.download.DownloadKeyViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/certificates/:name", app.download.DownloadCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/certrootchains/:name", app.download.DownloadCertRootChainViaHeader)

	// download keys and certs - via URL routes
	// include
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/privatekeys/:name/*apiKey", app.download.DownloadKeyViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/certificates/:name/*apiKey", app.download.DownloadCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/privatecerts/:name/*apiKey", app.download.DownloadPrivateCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiDownloadUrlPath+"/certrootchains/:name/*apiKey", app.download.DownloadCertRootChainViaUrl)

	// frontend (if enabled)
	if *app.config.ServeFrontend {
		// log availability
		app.logger.Infof("frontend hosting enabled and available at: %s", frontendUrlPath)

		// configure environment file
		app.setFrontendEnv()

		// redirect root to frontend app
		app.makeHandle(http.MethodGet, "/", redirectToFrontendRoot)

		// redirect base path to frontend app
		app.makeHandle(http.MethodGet, baseUrlPath, redirectToFrontendRoot)

		// add file server route for frontend
		app.makeHandle(http.MethodGet, frontendUrlPath+"/*anything", app.frontendHandler)
	}

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)
	app.router.MethodNotAllowed = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}
