package main

import "net/http"

func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := AppStatus{
		Status:      "Available",
		Environment: app.config.env,
		Version:     version,
	}

	app.WriteJSON(w, http.StatusOK, currentStatus, "status")

}
