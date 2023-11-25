package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// doShutdownHandler shuts LeGo down completely.
// Note: LeGo may still restart if the caller is configured to restart it
// if it stops (e.g. when running as a service).
func (app *Application) doShutdownHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// write response first since the action will shutdown server
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "lego shutting down"

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

// doRestartHandler shuts LeGo down and then calls the OS to execute LeGo
// again with the same args and environment as originally used.
func (app *Application) doRestartHandler(w http.ResponseWriter, r *http.Request) *output.Error {
	// write response first since the action will shutdown server
	response := &output.JsonResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "lego restarting"

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
