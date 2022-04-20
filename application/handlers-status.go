package application

import (
	"net/http"
)

func (app *Application) StatusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := AppStatus{
		Status:      "Available",
		Environment: app.Config.Env,
		Version:     version,
	}

	app.WriteJSON(w, http.StatusOK, currentStatus, "status")

}
