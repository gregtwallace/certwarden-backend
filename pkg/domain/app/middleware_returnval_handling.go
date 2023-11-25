package app

import (
	"fmt"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// loggableRequestURI returns a log safe RequestURI. The app baseUrlPath is removed
// and if the URI is a download URI, it is partially redacted if it contains an api
// key. For all other URIs, no special redaction is needed.
func loggableRequestURI(r *http.Request) string {
	uri := r.RequestURI

	// special handling if download URI
	if strings.HasPrefix(uri, apiKeyDownloadUrlPath) {
		// remove download prefix
		uri = strings.TrimPrefix(uri, apiKeyDownloadUrlPath)

		// split on /
		redactedURIPieces := strings.SplitN(uri, "/", 4)

		// include first two pieces (omit [0] because it is "" since original string starts with /)
		uri = "/" + redactedURIPieces[1] + "/" + redactedURIPieces[2]

		// add redacted 3rd, if it exists
		if len(redactedURIPieces) >= 4 {
			uri += "/" + output.RedactString(redactedURIPieces[3])
		}

		// add api key download prefix back
		uri = apiKeyDownloadUrlPath + uri
	}

	// always remove base path
	return strings.TrimPrefix(uri, baseUrlPath)
}

// middlewareApplyReturnValHandling applies middleware that transforms the custom handlerFunc into
// an http.Handler by processing the error from the custom handler func and logging it. If
// sensitive is true, the log level is increased and a little more verbose. This is useful
// for certain routes that should always log their access (e.g. download)
func middlewareApplyReturnValHandling(next handlerFunc, sensitive bool, logger *zap.SugaredLogger, output *output.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// shorten URI for logging
		trimmedURI := loggableRequestURI(r)

		// info log attempt if sensitive
		if sensitive {
			logger.Infof("client %s: attempting %s %s", r.RemoteAddr, r.Method, trimmedURI)
		}

		// execute
		start := time.Now()
		err := next(w, r)
		timeToServe := time.Since(start)

		// if there was an error, log it and write error JSON
		if err != nil {
			writeErr := output.WriteJSON(w, err)

			// if error, serve isn't done until error json written
			timeToServe = time.Since(start)

			if writeErr != nil {
				logger.Errorf("client %s: %s %s: failed to serve error response (json write error: %s)", r.RemoteAddr, r.Method, trimmedURI, writeErr)
			} else {
				logMsg := fmt.Sprintf("client %s: %s %s %v: served error response", r.RemoteAddr, r.Method, trimmedURI, timeToServe)
				if sensitive {
					logger.Info(logMsg)
				} else {
					logger.Debug(logMsg)
				}
			}

			// no error
		} else if sensitive {
			logger.Infof("client %s: %s %s %v: served without error", r.RemoteAddr, r.Method, trimmedURI, timeToServe)
		}
	}
}
