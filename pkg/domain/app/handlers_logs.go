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
					return output.ErrInternal
				}

				// create file in zip
				f, err := zipWriter.Create(name)
				if err != nil {
					app.logger.Error(err)
					return output.ErrInternal
				}

				// add file content to the zip file
				_, err = f.Write(dat)
				if err != nil {
					app.logger.Error(err)
					return output.ErrInternal
				}
			}
		}
	}

	// close zip writer
	err = zipWriter.Close()
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}

	// make zip filename with timestamp
	zipFilename := logFileName + "." + time.Now().Local().Format(time.RFC3339) + ".zip"

	// make data from byte buffer
	zipData := zipBuffer.Bytes()

	// output
	err = app.output.WriteZip(w, r, zipFilename, zipData)
	if err != nil {
		return err
	}

	return nil
}
