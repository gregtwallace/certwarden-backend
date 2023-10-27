package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// base path
const baseUrlPath = "/legocerthub"

// backend api paths
const apiUrlPath = baseUrlPath + "/api"
const apiKeyDownloadUrlPath = apiUrlPath + "/v1/download"

// frontend React app path (e.g. Vite config `base`)
const frontendUrlPath = baseUrlPath + "/app"

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) routes() http.Handler {
	app.router = httprouter.New()

	// health check (HEAD or GET) - insecure for docker probing
	app.handleAPIRouteInsecure(http.MethodHead, apiUrlPath+"/health", healthHandler)
	app.handleAPIRouteInsecure(http.MethodGet, apiUrlPath+"/health", healthHandler)

	// app auth - insecure as these give clients the access_token to access secure routes
	// validates with user/password
	app.handleAPIRouteInsecure(http.MethodPost, apiUrlPath+"/v1/app/auth/login", app.auth.LoginUsingUserPwPayload)
	// validates with cookie
	app.handleAPIRouteInsecure(http.MethodPost, apiUrlPath+"/v1/app/auth/refresh", app.auth.RefreshUsingCookie)

	// app auth - secure
	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/app/auth/changepassword", app.auth.ChangePassword)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/auth/logout", app.auth.Logout)

	// status
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/status", app.statusHandler)

	// app
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/log", app.viewCurrentLogHandler)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/logs", app.downloadLogsHandler)

	// app control
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/control/shutdown", app.doShutdownHandler)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/control/restart", app.doRestartHandler)

	// app updater
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/updater/new-version", app.updater.GetNewVersionInfo)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/updater/new-version", app.updater.CheckForNewVersion)

	// challenges (config)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/domains", app.challenges.Providers.GetAllDomains)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/services", app.challenges.Providers.GetAllProviders)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.Providers.GetOneProvider)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/challenges/providers/services", app.challenges.Providers.CreateProvider)
	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.Providers.ModifyProvider)
	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.Providers.DeleteProvider)

	// acme_servers
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeservers", app.acmeServers.GetAllServers)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.GetOneServer)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeservers", app.acmeServers.PostNewServer)
	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.PutServerUpdate)
	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.DeleteServer)

	// private_keys
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/privatekeys", app.keys.GetAllKeys)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id", app.keys.GetOneKey)
	app.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id/download", app.keys.DownloadOneKey)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/privatekeys", app.keys.PostNewKey)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.StageNewApiKey)
	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.RemoveOldApiKey)

	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/privatekeys/:id", app.keys.PutKeyUpdate)

	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)
	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/key-change", app.accounts.RolloverKey)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/new-account", app.accounts.NewAcmeAccount)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)

	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates", app.certificates.GetAllCerts)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates/:certid", app.certificates.GetOneCert)

	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates", app.certificates.PostNewCert)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.StageNewApiKey)
	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.RemoveOldApiKey)

	app.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/orders/fulfiller/status", app.orders.GetAllWorkStatus)

	app.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.NewOrder)

	app.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/download", app.orders.DownloadCertNewestOrder)
	app.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/download", app.orders.DownloadOneOrder)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatekeys/:name", app.download.DownloadKeyViaHeader)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certificates/:name", app.download.DownloadCertViaHeader)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certrootchains/:name", app.download.DownloadCertRootChainViaHeader)

	// download keys and certs - via URL routes
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatekeys/:name/*apiKey", app.download.DownloadKeyViaUrl)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certificates/:name/*apiKey", app.download.DownloadCertViaUrl)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecerts/:name/*apiKey", app.download.DownloadPrivateCertViaUrl)
	app.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certrootchains/:name/*apiKey", app.download.DownloadCertRootChainViaUrl)

	// frontend (if enabled)
	if *app.config.FrontendServe {
		// log availability
		app.logger.Infof("frontend hosting enabled and available at: %s", frontendUrlPath)

		// configure environment file
		setFrontendEnv(app.config.FrontendShowDebugInfo)

		// redirect root to frontend app
		app.handleFrontend(http.MethodGet, "/", redirectToFrontendHandler)

		// redirect base path to frontend app
		app.handleFrontend(http.MethodGet, baseUrlPath, redirectToFrontendHandler)

		// add file server route for frontend
		app.handleFrontend(http.MethodGet, frontendUrlPath+"/*anything(unused)", app.frontendHandler)
	}

	// invalid route
	app.router.NotFound = app.handlerNotFound()

	// options route
	app.router.HandleOPTIONS = true
	app.router.GlobalOPTIONS = app.handlerGlobalOptions()

	// wrong method
	app.router.HandleMethodNotAllowed = false

	return app.router
}
