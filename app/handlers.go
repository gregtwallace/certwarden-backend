package app

import (
	"legocerthub-backend/utils"
	"legocerthub-backend/utils/acme_utils"
	"net/http"
)

func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := appStatus{
		Status:      "Available",
		Environment: app.Config.Env,
		Version:     version,
		AcmeDirectories: appStatusDirectories{
			Production: acme_utils.LeProdDirectory,
			Staging:    acme_utils.LeStagingDirectory,
		},
	}

	utils.WriteJSON(w, http.StatusOK, currentStatus, "status")

}
