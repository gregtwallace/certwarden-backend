package app

import (
	"certwarden-backend/pkg/output"
	"net/http"
)

// doShutdownHandler shuts the app down completely.
// Note: It may still restart if the caller is configured to restart it
// if it stops (e.g. when running as a service).
func (app *Application) doShutdownHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// write response first since the action will shutdown server
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "shutting down"

	err := app.output.WriteJSON(w, response)
	if err != nil {
		app.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	// log shutdown
	app.logger.Infof("client %s: triggered graceful shutdown via api", r.RemoteAddr)

	// do shutdown
	app.shutdown(false)

	return nil
}

// doRestartHandler shuts the app down and then calls the OS to execute the app
// again with the same args and environment as originally used.
func (app *Application) doRestartHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// write response first since the action will shutdown server
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "restarting"

	err := app.output.WriteJSON(w, response)
	if err != nil {
		app.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	// log restart
	app.logger.Infof("client %s: triggered graceful restart via api", r.RemoteAddr)

	// do shutdown
	app.shutdown(true)

	return nil
}
