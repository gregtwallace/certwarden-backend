package app

import (
	"archive/zip"
	"bufio"
	"bytes"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// logEntriesForView is the number of log entries that should be returned in the
// log view API response
const logEntriesForView = 500

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

// readAndParseLogFile reads the specified log file and converts it into an array of log entries
func readAndParseLogFile(filePathAndName string) ([]logEntry, error) {
	// open log, read only
	logFile, err := os.OpenFile(filePathAndName, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()

	// read log line by line
	entries := []logEntry{}
	scanner := bufio.NewScanner(logFile)

	for scanner.Scan() {
		// read a line and remove the trailing comma
		entryLine := scanner.Bytes()
		entryLine = bytes.TrimSuffix(entryLine, []byte{','})

		entry := logEntry{}
		err = json.Unmarshal(entryLine, &entry)
		if err != nil {
			// if line failed to unmarshal, add an entry describing the corruption
			entry = logEntry{
				// use linux epoch time as garbage value since most users should recognize it
				// as unix epoch, and thus assume it is probably nonsense
				TimeStamp: time.Unix(0, 0).Format("2006-01-02T15:04:05.000Z0700"),
				Level:     "error",
				Caller:    "unknown",
				Message:   "failed to parse log line (log file line is corrupt)",
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// viewLogHandler is a handler that returns the content of the current log file to the client
func (app *Application) viewCurrentLogHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// open log, read only
	logEntries, err := readAndParseLogFile(dataStorageLogPath + "/" + logFileName)
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// if size is too small, read more from next file
	if len(logEntries) < logEntriesForView {
		logFileNames, err := listLogFiles()
		if err != nil {
			// log error but don't fail the call since the first log file parsed fine
			app.logger.Errorf("app: log view: failed to list log files (%s)", err)
		} else {
			oldestTimestampLogFilename := ""
			newestTime := time.Time{}

			for _, logFilename := range logFileNames {
				// skip the in-use log file
				if logFilename == logFileName {
					continue
				}

				// parse the time from each filename and keep the newest one (which would be the 2nd newest file since
				// active file doesn't have a date string)
				timeString := strings.TrimSuffix(strings.TrimPrefix(logFilename, logFileBaseName+"-"), logFileSuffix)

				t, err := time.Parse("2006-01-02T15-04-05.000", timeString)
				if err != nil {
					app.logger.Errorf("app: log view: failed to parse time of log file %s (%s)", logFilename, err)
				} else {
					if t.After(newestTime) {
						newestTime = t
						oldestTimestampLogFilename = logFilename
					}
				}
			}

			if oldestTimestampLogFilename != "" {
				olderLogEntries, err := readAndParseLogFile(dataStorageLogPath + "/" + oldestTimestampLogFilename)
				if err != nil {
					// log error but don't fail the call since the first log file parsed fine
					app.logger.Errorf("app: log view: failed to read next oldest log file %s (%s)", oldestTimestampLogFilename, err)
				} else {
					logEntries = append(olderLogEntries, logEntries...)
				}
			}
		}
	}

	// if size is too big, truncate
	if len(logEntries) > logEntriesForView {
		logEntries = logEntries[len(logEntries)-logEntriesForView:]
	}

	// write response
	response := &currentLogResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.LogEntries = logEntries

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

	// get all log files
	logFiles, err := listLogFiles()
	if err != nil {
		app.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// range through all log files
	for _, logFilename := range logFiles {
		// open log file
		logFile, err := os.Open(dataStorageLogPath + "/" + logFilename)
		if err != nil {
			app.logger.Error(err)
			return output.JsonErrInternal(err)
		}
		defer logFile.Close()

		// create file in zip
		zipFile, err := zipWriter.Create(logFilename)
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
