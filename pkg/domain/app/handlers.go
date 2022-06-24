package app

import (
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
