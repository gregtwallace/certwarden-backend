package app

import (
	"archive/zip"
	"bytes"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// logEntry represents the structure of the zap log
type logEntry struct {
	Level      string `json:"level"`
	TimeStamp  string `json:"ts"`
	Caller     string `json:"caller"`
	Message    string `json:"msg"`
	StackTrace string `json:"stacktrace,omitempty"`
}

type currentLogResponse struct {
	output.JsonResponse
	LogEntries []logEntry `json:"log_entries"`
}

// viewLogHandler is a handler that returns the content of the current log file to the client
func (app *Application) viewCurrentLogHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// open log, read only
	logFile, err := os.OpenFile(dataStorageLogPath+"/"+logFileName, os.O_RDONLY, 0600)
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}
	defer logFile.Close()

	// read in the log file (with slight modifications to make valid json array)
	logBuffer := &bytes.Buffer{}
	// add opening wrap and bracket for json
	_, _ = logBuffer.WriteString("[")
	_, _ = io.Copy(logBuffer, logFile)
	// remove line break and comma after last log item
	logBuffer.Truncate(logBuffer.Len() - 2)
	// add end of json
	_, _ = logBuffer.WriteString("]")

	// Unmarshal the log entries in response
	response := &currentLogResponse{}
	err = json.Unmarshal(logBuffer.Bytes(), &response.LogEntries)
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	// response.LogEntries = [populated above]

	// return response to client
	err = app.output.WriteJSON(w, response)
	if err != nil {
		app.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}

// downloadLogsHandler is a handler that sends a zip of all of the log files to
// the client
func (app *Application) downloadLogsHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// make buffer and writer for zip
	zipBuffer := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(zipBuffer)

	// get all files in the log directory
	files, err := os.ReadDir(dataStorageLogPath)
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// range all files in log directory
	for i := range files {
		// ignore directories
		if files[i].IsDir() {
			continue
		}

		name := files[i].Name()

		// confirm prefix and suffix then add (aka ensure non-log files that are accidentally in
		// this folder are not zipped up and returned to client)
		// also check for old log file names (pre- app rename)
		if !((strings.HasPrefix(name, logFileBaseName) || strings.HasPrefix(name, "lego-certhub")) &&
			strings.HasSuffix(name, logFileSuffix)) {

			continue
		}

		// open log file
		logFile, err := os.Open(dataStorageLogPath + "/" + name)
		if err != nil {
			app.logger.Error(err)
			return output.JsonErrInternal(err)
		}
		defer logFile.Close()

		// create file in zip
		zipFile, err := zipWriter.Create(name)
		if err != nil {
			app.logger.Error(err)
			return output.JsonErrInternal(err)
		}

		// copy log file to zip file
		_, err = io.Copy(zipFile, logFile)
		if err != nil {
			app.logger.Error(err)
			return output.JsonErrInternal(err)
		}
	}

	// close zip writer (note: Close() writes the gzip footer and cannot be deferred)
	err = zipWriter.Close()
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// make zip filename with timestamp
	zipFilenameNoExt := logFileName + "." + time.Now().Local().Format(time.RFC3339)

	// output
	app.output.WriteZip(w, r, zipFilenameNoExt, zipBuffer.Bytes())

	return nil
}
