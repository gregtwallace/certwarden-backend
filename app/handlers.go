package app

import (
	"legocerthub-backend/utils"
	"net/http"
)

func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := appStatus{
		Status:      "Available",
		Environment: app.Config.Env,
		Version:     version,
	}

	utils.WriteJSON(w, http.StatusOK, currentStatus, "status")

}
