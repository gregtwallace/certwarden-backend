package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// doShutdownHandler triggers LeGo to shutdown
func (app *Application) doShutdownHandler(w http.ResponseWriter, r *http.Request) (err error) {

	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "lego shutdown triggered",
	}

	_, err = app.output.WriteJSON(w, http.StatusOK, response, "response")
	if err != nil {
		return err
	}

	app.logger.Infow("client %s triggered graceful shutdown via api", r.RemoteAddr)
	app.shutdown(false)

	return nil
}

// TODO: Enable - see comments in routes.go
// doRestartHandler triggers LeGo to restart
// func (app *Application) doRestartHandler(w http.ResponseWriter, r *http.Request) (err error) {

// 	response := output.JsonResponse{
// 		Status:  http.StatusOK,
// 		Message: "lego restart triggered",
// 	}

// 	_, err = app.output.WriteJSON(w, http.StatusOK, response, "response")
// 	if err != nil {
// 		return err
// 	}

// 	app.logger.Infow("client %s triggered graceful restart via api", r.RemoteAddr)
// 	app.shutdown(true)

// 	return nil
// }
