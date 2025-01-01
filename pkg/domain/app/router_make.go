package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// base path
const baseUrlPath = "/certwarden"

// backend api paths
const apiUrlPath = baseUrlPath + "/api"
const apiKeyDownloadUrlPath = apiUrlPath + "/v1/download"

// frontend React app path (e.g. Vite config `base`)
const frontendUrlPath = baseUrlPath + "/app"

// makeRouterAndRoutes creates the application's router and adds the routes. It also
// inserts the common CORS middleware before assigning the router to app
func (app *Application) makeRouterAndRoutes() {
	router := &router{
		logger:                app.logger.SugaredLogger,
		output:                app.output,
		auth:                  app.auth,
		permittedCrossOrigins: app.config.CORSPermittedCrossOrigins,
		r:                     httprouter.New(),
	}

	// health check (HEAD or GET) - insecure for docker probing
	router.handleAPIRouteInsecure(http.MethodHead, apiUrlPath+"/health", healthHandler)
	router.handleAPIRouteInsecure(http.MethodGet, apiUrlPath+"/health", healthHandler)

	// app auth - insecure as these give clients the access_token to access secure routes
	// app auth status
	router.handleAPIRouteInsecure(http.MethodGet, apiUrlPath+"/v1/app/auth/status", app.auth.Status)
	// validates with user/password
	router.handleAPIRouteInsecure(http.MethodPost, apiUrlPath+"/v1/app/auth/login", app.auth.LocalPostLogin)
	// validates with cookie
	router.handleAPIRouteInsecure(http.MethodPost, apiUrlPath+"/v1/app/auth/refresh", app.auth.RefreshUsingCookie)

	// app auth - secure
	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/app/auth/changepassword", app.auth.ChangePassword)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/auth/logout", app.auth.Logout)

	// status
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/status", app.statusHandler)

	// app
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/log", app.viewCurrentLogHandler)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/logs", app.downloadLogsHandler)

	// app control
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/control/shutdown", app.doShutdownHandler)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/control/restart", app.doRestartHandler)

	// app updater
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/updater/new-version", app.updater.GetNewVersionInfo)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/updater/new-version", app.updater.CheckForNewVersion)

	// app backup and restore
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/backup/disk", app.backup.ListDiskBackupsHandler)

	router.handleAPIRouteSecureSensitive(http.MethodPost, apiUrlPath+"/v1/app/backup/disk", app.backup.MakeDiskBackupNowHandler)
	router.handleAPIRouteSecureSensitive(http.MethodDelete, apiUrlPath+"/v1/app/backup/disk/:filename", app.backup.DeleteDiskBackupHandler)

	router.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/app/backup", app.backup.DownloadBackupNowHandler)
	router.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/app/backup/disk/:filename", app.backup.DownloadDiskBackupHandler)

	// challenges: dns aliases
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/domainaliases", app.challenges.GetDomainAliases)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/challenges/domainaliases", app.challenges.PostDomainAliases)

	// challenges: providers
	// router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/domains", app.challenges.Providers.GetAllDomains)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/services", app.challenges.DNSIdentifierProviders.GetAllProviders)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.DNSIdentifierProviders.GetOneProvider)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/app/challenges/providers/services", app.challenges.DNSIdentifierProviders.CreateProvider)
	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.DNSIdentifierProviders.ModifyProvider)
	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/app/challenges/providers/services/:id", app.challenges.DNSIdentifierProviders.DeleteProvider)

	// acme_servers
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeservers", app.acmeServers.GetAllServers)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.GetOneServer)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeservers", app.acmeServers.PostNewServer)
	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.PutServerUpdate)
	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/acmeservers/:id", app.acmeServers.DeleteServer)

	// private_keys
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/privatekeys", app.keys.GetAllKeys)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id", app.keys.GetOneKey)
	router.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id/download", app.keys.DownloadOneKey)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/privatekeys", app.keys.PostNewKey)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.StageNewApiKey)
	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.RemoveOldApiKey)

	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/privatekeys/:id", app.keys.PutKeyUpdate)

	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeaccounts", app.accounts.GetAllAccounts)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts", app.accounts.PostNewAccount)

	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)
	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/key-change", app.accounts.RolloverKey)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/register", app.accounts.NewAcmeAccount)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/refresh", app.accounts.RefreshAcmeAccount)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/post-as-get", app.accounts.PostAsGet)

	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates", app.certificates.GetAllCerts)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates/:certid", app.certificates.GetOneCert)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates", app.certificates.PostNewCert)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.StageNewApiKey)
	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.RemoveOldApiKey)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/clientkey", app.certificates.MakeNewClientKey)
	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/clientkey", app.certificates.DisableClientKey)

	router.handleAPIRouteSecure(http.MethodPut, apiUrlPath+"/v1/certificates/:certid", app.certificates.PutDetailsCert)

	router.handleAPIRouteSecure(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/orders/fulfilling/status", app.orders.GetFulfillWorkStatus)
	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/orders/post-process/status", app.orders.GetPostProcessWorkStatus)

	router.handleAPIRouteSecure(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.NewOrder)

	router.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/download", app.orders.DownloadCertNewestOrder)
	router.handleAPIRouteSecureDownload(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/download", app.orders.DownloadOneOrder)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	router.handleAPIRouteSecure(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/postprocess", app.orders.PostProcessOrder)

	// download keys and certs
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatekeys/:name", app.download.DownloadKeyViaHeader)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certificates/:name", app.download.DownloadCertViaHeader)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecertchains/:name", app.download.DownloadPrivateCertChainViaHeader)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certrootchains/:name", app.download.DownloadCertRootChainViaHeader)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/pfx/:name", app.download.DownloadPfxViaHeader)

	// download keys and certs - via URL routes
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatekeys/:name/*apiKey", app.download.DownloadKeyViaUrl)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certificates/:name/*apiKey", app.download.DownloadCertViaUrl)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecerts/:name/*apiKey", app.download.DownloadPrivateCertViaUrl)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/privatecertchains/:name/*apiKey", app.download.DownloadPrivateCertChainViaUrl)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/certrootchains/:name/*apiKey", app.download.DownloadCertRootChainViaUrl)
	router.handleAPIRouteDownloadWithAPIKey(http.MethodGet, apiKeyDownloadUrlPath+"/pfx/:name/*apiKey", app.download.DownloadPfxViaUrl)

	// frontend (if enabled)
	if *app.config.FrontendServe {
		// log availability
		app.logger.Infof("frontend hosting enabled and available at: %s", frontendUrlPath)

		// configure environment file
		err := setFrontendEnv(app.config.FrontendShowDebugInfo)
		if err != nil {
			// don't fail, just error log the problem (sane default values should be in place)
			app.logger.Errorf("frontend: failed to set frontend environment (%s)", err)
		}

		// redirect root to frontend app
		router.handleFrontend(http.MethodGet, "/", redirectToFrontendHandler)

		// redirect base path to frontend app
		router.handleFrontend(http.MethodGet, baseUrlPath, redirectToFrontendHandler)

		// add file server route for frontend
		router.handleFrontend(http.MethodGet, frontendUrlPath+"/*anything(unused)", app.frontendFileHandler)
	}

	// invalid route
	router.r.NotFound = app.handlerNotFound()

	// options route
	router.r.HandleOPTIONS = true
	router.r.GlobalOPTIONS = app.handlerGlobalOptions()

	// wrong method
	router.r.HandleMethodNotAllowed = false

	// convert to handler and apply common (universal) middlewares
	appRouter := http.Handler(router)
	// browser security headers (for all routes, not just frontend)
	appRouter = middlewareApplyBrowserSecurityHeaders(appRouter)
	// HSTS header (only is HTTPS and config option to disable isn't true)
	if app.IsHttps() && (app.config.DisableHSTS != nil || !*app.config.DisableHSTS) {
		appRouter = middlewareApplyHSTS(appRouter)
	}

	// set app's router
	app.router = appRouter
}
