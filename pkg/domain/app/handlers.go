package app

import (
	"encoding/json"
	"io"
	"legocerthub-backend/pkg/output"
	"net/http"
	"os"
)

type appStatus struct {
	Status          string               `json:"status"`
	DevMode         bool                 `json:"development_mode,omitempty"`
	Version         string               `json:"version"`
	ApiUrl          string               `json:"api_url"`
	FrontendUrl     string               `json:"frontend_url,omitempty"`
	AcmeDirectories appStatusDirectories `json:"acme_directories"`
}

type appStatusDirectories struct {
	Production string `json:"prod"`
	Staging    string `json:"staging"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) (err error) {

	currentStatus := appStatus{
		Status:      "available",
		DevMode:     *app.config.DevMode,
		Version:     appVersion,
		ApiUrl:      app.ApiUrl(),
		FrontendUrl: app.FrontendUrl(),
		AcmeDirectories: appStatusDirectories{
			Production: acmeProdUrl,
			Staging:    acmeStagingUrl,
		},
	}

	_, err = app.output.WriteJSON(w, http.StatusOK, currentStatus, "server")
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// logEntry represents the structure of the zap log
type logEntry struct {
	Level      string `json:"level"`
	TimeStamp  string `json:"ts"`
	Caller     string `json:"caller"`
	Message    string `json:"msg"`
	StackTrace string `json:"stacktrace,omitempty"`
}

// viewLogHandler is a handler that returns the content of the log file to the client
func (app *Application) viewLogHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// open log, read only
	logFile, err := os.OpenFile(logFile, os.O_RDONLY, 0600)
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}

	// read in the log file
	logBytes, err := io.ReadAll(logFile)
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}
	// manipulate the log to make it proper json
	jsonBytes := []byte("[")
	jsonBytes = append(jsonBytes, logBytes...)
	jsonBytes = append(jsonBytes, []byte("{}]")...) // add empty item because final log ends in ,

	// unmarshal to cleanup the json
	var logsJson []logEntry
	err = json.Unmarshal(jsonBytes, &logsJson)
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}

	// output
	// Note: len -1 is to remove the last empty item
	_, err = app.output.WriteJSON(w, http.StatusOK, logsJson[:len(logsJson)-1], "logs")
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// notFoundHandler is called when there is not a matching route on the router. If in dev,
// return 404 so issue is clear. If in prod, return the generic 401 unauthorized to prevent
// unauthorized parties from checking what routes exist. This is definitely overkill since
// the software is open source.
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// default error 401
	outError := output.ErrUnauthorized

	// if devMode, return 404
	if app.GetDevMode() {
		outError = output.ErrNotFound
	}

	_, err = app.output.WriteErrorJSON(w, outError)
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}
