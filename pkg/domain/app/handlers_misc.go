package app

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/storage"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// serverStatusResponse
type serverStatusResponse struct {
	output.JsonResponse

	ServerStatus struct {
		Status        string `json:"status"`
		LogLevel      string `json:"log_level"`
		Version       string `json:"version"`
		ConfigVersion int    `json:"config_version"`
		DbUserVersion int    `json:"database_version"`
	} `json:"server"`
}

// statusHandler writes some basic info about the status of the Application
func (app *Application) statusHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// write response
	response := &serverStatusResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.ServerStatus.Status = "available"
	response.ServerStatus.LogLevel = app.logger.Level().String()
	response.ServerStatus.Version = appVersion
	response.ServerStatus.ConfigVersion = *app.config.ConfigVersion
	response.ServerStatus.DbUserVersion = storage.DbCurrentUserVersion

	err := app.output.WriteJSON(w, response)
	if err != nil {
		app.logger.Errorf("failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}

	return nil
}

// healthHandler writes some basic info about the status of the Application
func healthHandler(w http.ResponseWriter, r *http.Request) *output.JsonError {
	// write 204 (No Content)
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// httpToHttpsRedirectHandler redirects the user from http to https. If the redirect target isn't
// in the https certificate, 400 is returned instead
func (app *Application) httpToHttpsRedirectHandler(w http.ResponseWriter, r *http.Request) {
	// get hostname from http request
	hostName, _, err := net.SplitHostPort(r.Host)
	isRawV6 := false
	if err != nil {
		// if it failed, assume there is no port and use raw r.Host
		hostName = r.Host
		if len(hostName) > 1 && hostName[0] == '[' && hostName[len(hostName)-1:][0] == ']' {
			hostName = hostName[1 : len(hostName)-1]
			isRawV6 = true
		}
	}

	// use https cert as the whitelist of permitted hostnames
	if !app.httpsCert.ContainsHostname(hostName) {
		// Invalid Redirect target -- write 400 (Bad Request)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		//nolint:errcheck // don't care if this doesn't write, user errored out anyway
		if isRawV6 {
			hostName = "[" + hostName + "]"
		}
		fmt.Fprintf(w, "redirect failed: https hostname '%s' is not a part of the server's certificate", hostName)
		return
	}

	// build redirect address (clunky because of the way httptest sets `RequestURI`)
	if isRawV6 {
		hostName = "[" + hostName + "]"
	}
	newAddr := "https://" + hostName + ":" + strconv.Itoa(*app.config.HttpsPort)
	if r.URL.Path != "" {
		newAddr += "/" + strings.TrimPrefix(r.URL.Path, "/")
	}
	if r.URL.RawQuery != "" {
		newAddr += "?" + strings.TrimPrefix(r.URL.RawQuery, "?")
	}

	http.Redirect(w, r, newAddr, http.StatusTemporaryRedirect)
}
