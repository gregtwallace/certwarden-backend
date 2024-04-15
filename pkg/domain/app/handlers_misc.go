package app

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage/sqlite"
	"net/http"
	"strings"
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
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) *output.Error {
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
		return output.ErrWriteJsonError
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func healthHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// write 204 (No Content)
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// redirectOldLeGoName is a handler to redirect old LeGo routes to the new ones
// TODO: Remove eventually.
func (app *Application) redirectOldLeGoName(w http.ResponseWriter, r *http.Request) {
	// if this isn't the old LeGo base path, throw an error (should never happen but
	// just in case to stop any infinite redirect)
	pathSuffix, hasPrefix := strings.CutPrefix(r.URL.Path, "/legocerthub")
	if !hasPrefix {
		app.logger.Errorf("client: %s: redirect error (wrong base path)", r.RemoteAddr)
	}
	newPath := baseUrlPath + pathSuffix

	// log warning so user knows clients that need to be updated
	app.logger.Warnf("client: %s: old base path (pre- app rename) redirected; update client asap", r.RemoteAddr)

	http.Redirect(w, r, newPath, http.StatusMovedPermanently)
}
