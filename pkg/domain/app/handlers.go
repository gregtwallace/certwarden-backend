package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/storage/sqlite"
	"net/http"
	"os"
	"strings"
	"time"
)

type appStatus struct {
	Status             string `json:"status"`
	DevMode            bool   `json:"development_mode,omitempty"`
	Version            string `json:"version"`
	DbUserVersion      int    `json:"database_version"`
	ConfigVersionMatch bool   `json:"config_version_match"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) (err error) {

	currentStatus := appStatus{
		Status:             "available",
		DevMode:            *app.config.DevMode,
		Version:            appVersion,
		DbUserVersion:      sqlite.DbCurrentUserVersion,
		ConfigVersionMatch: app.config.ConfigVersion == configVersion,
	}

	_, err = app.output.WriteJSON(w, http.StatusOK, currentStatus, "server")
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func (app *Application) healthHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// write 204 (No Content)
	app.output.WriteEmptyResponse(w, http.StatusNoContent)

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

// viewLogHandler is a handler that returns the content of the current log file to the client
func (app *Application) viewCurrentLogHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// open log, read only
	logFile, err := os.OpenFile(logFilePath+logFileName, os.O_RDONLY, 0600)
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
	_, err = app.output.WriteJSON(w, http.StatusOK, logsJson[:len(logsJson)-1], "log_entries")
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// downloadLogsHandler is a handler that sends a zip of all of the log files to
// the client
func (app *Application) downloadLogsHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// make buffer and writer for zip
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// get all files in the log directory
	files, err := os.ReadDir(logFilePath)
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteZipFailed
	}

	// for each file in the log dir, verify start and end of filename then add it to the zip
	logFileSplit := strings.Split(logFileName, ".")
	logFilePrefix := logFileSplit[0]
	logFileSuffix := logFileSplit[len(logFileSplit)-1]

	// range all files in log directory
	for i := range files {
		// ignore directories
		if !files[i].IsDir() {
			name := files[i].Name()

			// confirm prefix and suffix then add
			if strings.HasPrefix(name, logFilePrefix) && strings.HasSuffix(name, logFileSuffix) {

				// read file content
				dat, err := os.ReadFile(logFilePath + name)
				if err != nil {
					app.logger.Error(err)
					return output.ErrWriteZipFailed
				}

				// create file in zip
				f, err := zipWriter.Create(name)
				if err != nil {
					app.logger.Error(err)
					return output.ErrWriteZipFailed
				}

				// add file content to the zip file
				_, err = f.Write(dat)
				if err != nil {
					app.logger.Error(err)
					return output.ErrWriteZipFailed
				}
			}
		}
	}

	// close zip writer
	err = zipWriter.Close()
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteZipFailed
	}

	// make zip filename with timestamp
	zipFilename := logFileName + "." + time.Now().Local().Format(time.RFC3339) + ".zip"

	// output
	_, err = app.output.WriteZip(w, zipFilename, zipBuffer)
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteZipFailed
	}

	return nil
}

// notFoundHandler is called when there is not a matching route on the router. If in dev,
// return 404 so issue is clear. If in prod, return the generic 401 unauthorized to prevent
// unauthorized parties from checking what routes exist. This is definitely overkill since
// the software is open source.
func (app *Application) notFoundHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// if OPTIONS, return no content
	// otherwise, preflight errors occur for bad routes (see: https://stackoverflow.com/questions/52047548/response-for-preflight-does-not-have-http-ok-status-in-angular)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// default error 401
	outError := output.ErrUnauthorized

	// if devMode, return 404
	if app.GetDevMode() {
		outError = output.ErrNotFound
	}

	// return error
	_, err = app.output.WriteErrorJSON(w, outError)
	if err != nil {
		app.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// redirectAddBasePathHandler adds the baseUrlPath to the request and then redirects
// TODO: Remove eventually?
func redirectAddBasePathHandler(w http.ResponseWriter, r *http.Request) error {
	// add base path
	newPath := baseUrlPath + r.URL.Path

	http.Redirect(w, r, newPath, http.StatusPermanentRedirect)
	return nil
}

// redirectToFrontendRoot is a handler that redirects to the frontend app
func redirectToFrontendRoot(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, frontendUrlPath, http.StatusPermanentRedirect)
	return nil
}
