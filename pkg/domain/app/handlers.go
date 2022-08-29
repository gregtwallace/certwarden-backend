package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

type appStatus struct {
	Status          string               `json:"status"`
	DevMode         bool                 `json:"development_mode,omitempty"`
	Version         string               `json:"version"`
	AcmeDirectories appStatusDirectories `json:"acme_directories"`
}

type appStatusDirectories struct {
	Production string `json:"prod"`
	Staging    string `json:"staging"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) (err error) {

	currentStatus := appStatus{
		Status:  "available",
		DevMode: *app.config.DevMode,
		Version: appVersion,
		AcmeDirectories: appStatusDirectories{
			Production: acmeProdUrl,
			Staging:    acmeStagingUrl,
		},
	}

	_, err = app.output.WriteJSON(w, http.StatusOK, currentStatus, "server")
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// notFoundHandler is called when there is not a matching route on the router. If in dev,
// return 404 so issue is clear. If in prod, return the generic 401 unauthorized to prevent
// unauthorized parties from checking what routes exist. This is definitely overkill since
// the software is open source.
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// default error 401
	outError := output.ErrUnauthorized

	// if devMode, return 404
	if app.GetDevMode() {
		outError = output.ErrNotFound
	}

	_, err = app.output.WriteErrorJSON(w, outError)
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
