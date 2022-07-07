package app

import (
	"errors"
	"legocerthub-backend/pkg/utils"
	"net/http"
)

type appStatus struct {
	Status          string               `json:"status"`
	Environment     string               `json:"environment"`
	Version         string               `json:"version"`
	AcmeDirectories appStatusDirectories `json:"acme_directories"`
}

type appStatusDirectories struct {
	Production string `json:"prod"`
	Staging    string `json:"staging"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) {

	currentStatus := appStatus{
		Status:  "Available",
		Version: version,
		AcmeDirectories: appStatusDirectories{
			Production: acmeProdUrl,
			Staging:    acmeStagingUrl,
		},
	}

	utils.WriteJSON(w, http.StatusOK, currentStatus, "status")
}

// notFoundHandler is called when there is not a matching route on the router
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	utils.WriteErrorStatusJSON(w, http.StatusNotFound, errors.New("route not found"))
}
