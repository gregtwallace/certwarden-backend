package app

import (
	"archive/zip"
	"bytes"
	"io"
	"legocerthub-backend/pkg/output"
	"net/http"
	"os"
	"strings"
	"time"
)

// logEntry represents the structure of the zap log
// type logEntry struct {
// 	Level      string `json:"level"`
// 	TimeStamp  string `json:"ts"`
// 	Caller     string `json:"caller"`
// 	Message    string `json:"msg"`
// 	StackTrace string `json:"stacktrace,omitempty"`
// }

// viewLogHandler is a handler that returns the content of the current log file to the client
func (app *Application) viewCurrentLogHandler(w http.ResponseWriter, r *http.Request) (err error) {
	// open log, read only
	logFile, err := os.OpenFile(logFilePath+logFileName, os.O_RDONLY, 0600)
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}
	defer logFile.Close()

	// read in the log file
	logBuffer := &bytes.Buffer{}
	// add opening wrap and bracket for json
	_, _ = logBuffer.WriteString("{\"log_entries\": [")
	_, _ = io.Copy(logBuffer, logFile)
	// remove line break and comma after last log item
	logBuffer.Truncate(logBuffer.Len() - 2)
	// add end of json
	_, _ = logBuffer.WriteString("]}")

	// output
	err = app.output.WriteMarshalledJSON(w, http.StatusOK, logBuffer.Bytes())
	if err != nil {
		return err
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
		return output.ErrInternal
	}

	// range all files in log directory
	for i := range files {
		// ignore directories
		if !files[i].IsDir() {
			name := files[i].Name()

			// confirm prefix and suffix then add
			if strings.HasPrefix(name, logFileBaseName) && strings.HasSuffix(name, logFileSuffix) {

				// open log file
				logFile, err := os.Open(logFilePath + name)
				if err != nil {
					app.logger.Error(err)
					return output.ErrInternal
				}
				defer logFile.Close()

				// create file in zip
				zipFile, err := zipWriter.Create(name)
				if err != nil {
					app.logger.Error(err)
					return output.ErrInternal
				}

				// copy log file to zip file
				_, err = io.Copy(zipFile, logFile)
				if err != nil {
					app.logger.Error(err)
					return output.ErrInternal
				}
			}
		}
	}

	// close zip writer (note: Close() writes the gzip footer and cannot be deferred)
	err = zipWriter.Close()
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}

	// make zip filename with timestamp
	zipFilename := logFileName + "." + time.Now().Local().Format(time.RFC3339) + ".zip"

	// output
	err = app.output.WriteZip(w, r, zipFilename, zipBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
