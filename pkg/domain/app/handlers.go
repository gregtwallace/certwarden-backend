package app

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
)

// statusHandler writes some basic info as json
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := appStatus{
		Status:  "Available",
		Version: version,
		AcmeDirectories: appStatusDirectories{
			Production: app.acmeProd.TosUrl(),
			Staging:    app.acmeStaging.TosUrl(),
		},
	}

	utils.WriteJSON(w, http.StatusOK, currentStatus, "status")
}

// notFoundHandler is called when there is not a matching route on the router
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteErrorStatusJSON(w, http.StatusNotFound, errors.New("route not found"))
}
