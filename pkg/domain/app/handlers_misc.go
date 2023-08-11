package app

import (
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
)

type appStatus struct {
	Status             string `json:"status"`
	DevMode            bool   `json:"development_mode,omitempty"`
	Version            string `json:"version"`
	DbUserVersion      int    `json:"database_version"`
	ConfigVersionMatch bool   `json:"config_version_match"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) (err error) {

	currentStatus := appStatus{
		Status:             "available",
		DevMode:            *app.config.DevMode,
		Version:            appVersion,
		DbUserVersion:      sqlite.DbCurrentUserVersion,
		ConfigVersionMatch: app.config.ConfigVersion == configVersion,
	}

	err = app.output.WriteJSON(w, http.StatusOK, currentStatus, "server")
	if err != nil {
		return err
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func (app *Application) healthHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// write 204 (No Content)
	app.output.WriteEmptyResponse(w, http.StatusNoContent)

	return nil
}

// notFoundHandler is called when there is not a matching route on the router. If in dev,
// return 404 so issue is clear. If in prod, return the generic 401 unauthorized to prevent
// unauthorized parties from checking what routes exist. This is definitely overkill since
// the software is open source.
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// if OPTIONS, return no content
	// otherwise, preflight errors occur for bad routes (see: https://stackoverflow.com/questions/52047548/response-for-preflight-does-not-have-http-ok-status-in-angular)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// default error 401
	outError := output.ErrUnauthorized

	// if devMode, return 404
	if app.GetDevMode() {
		outError = output.ErrNotFound
	}

	// return error
	err = app.output.WriteErrorJSON(w, outError)
	if err != nil {
		return err
	}

	return nil
}

// redirectAddBasePathHandler adds the baseUrlPath to the request and then redirects
// TODO: Remove eventually?
func redirectAddBasePathHandler(w http.ResponseWriter, r *http.Request) error {
	// add base path
	newPath := baseUrlPath + r.URL.Path

	http.Redirect(w, r, newPath, http.StatusPermanentRedirect)
	return nil
}

// redirectToFrontendRoot is a handler that redirects to the frontend app
func redirectToFrontendRoot(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, frontendUrlPath, http.StatusPermanentRedirect)
	return nil
}
