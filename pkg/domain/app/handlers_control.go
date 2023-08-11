package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// doShutdownHandler shuts LeGo down completely.
// Note: LeGo may still restart if the caller is configured to restart it
// if it stops (e.g. when running as a service).
func (app *Application) doShutdownHandler(w http.ResponseWriter, r *http.Request) (err error) {
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "lego shutdown triggered",
	}

	err = app.output.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		return err
	}

	app.logger.Infof("client %s triggered graceful shutdown via api", r.RemoteAddr)
	app.shutdown(false)

	return nil
}

// doRestartHandler shuts LeGo down and then calls the OS to execute LeGo
// again with the same args and environment as originally used.
func (app *Application) doRestartHandler(w http.ResponseWriter, r *http.Request) (err error) {
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "lego restart triggered",
	}

	err = app.output.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		return err
	}

	app.logger.Infof("client %s triggered graceful restart via api", r.RemoteAddr)
	app.shutdown(true)

	return nil
}
