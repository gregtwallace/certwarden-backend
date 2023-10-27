package app

import (
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
)

type appStatus struct {
	Status        string `json:"status"`
	LogLevel      string `json:"log_level"`
	Version       string `json:"version"`
	ConfigVersion int    `json:"config_version"`
	DbUserVersion int    `json:"database_version"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) (err error) {
	currentStatus := appStatus{
		Status:        "available",
		LogLevel:      app.logger.Level().String(),
		Version:       appVersion,
		ConfigVersion: *app.config.ConfigVersion,
		DbUserVersion: sqlite.DbCurrentUserVersion,
	}

	err = app.output.WriteJSON(w, http.StatusOK, currentStatus, "server")
	if err != nil {
		return err
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func healthHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// write 204 (No Content)
	w.WriteHeader(http.StatusNoContent)

	return nil
}
