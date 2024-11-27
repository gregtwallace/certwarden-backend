package app

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage/sqlite"
	"net/http"
)

// serverStatusResponse
type serverStatusResponse struct {
	output.JsonResponse
	ServerStatus struct {
		Status        string `json:"status"`
		LogLevel      string `json:"log_level"`
		Version       string `json:"version"`
		ConfigVersion int    `json:"config_version"`
		DbUserVersion int    `json:"database_version"`
	} `json:"server"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// write response
	response := &serverStatusResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.ServerStatus.Status = "available"
	response.ServerStatus.LogLevel = app.logger.Level().String()
	response.ServerStatus.Version = appVersion
	response.ServerStatus.ConfigVersion = *app.config.ConfigVersion
	response.ServerStatus.DbUserVersion = sqlite.DbCurrentUserVersion

	err := app.output.WriteJSON(w, response)
	if err != nil {
		app.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func healthHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// write 204 (No Content)
	w.WriteHeader(http.StatusNoContent)

	return nil
}
