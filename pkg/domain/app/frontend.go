package app

import (
	"bytes"
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const frontendBuildDir = "./frontend_build"
const frontendEnvFile = frontendBuildDir + "/env.js"

// noncePlaceholder is the text to use in frontend to show server where to inject nonce
var noncePlaceholder = []byte("{SERVER-CSP-NONCE}")

// setContentSecurityPolicy sets w's CSP to allow a very limited subset of content that the
// LeGo react app loads.
func setContentSecurityPolicy(w http.ResponseWriter, nonce []byte) {
	// LeGo app's security policy
	nonceString := string(nonce)
	var contentSecurityPolicy = []string{
		"default-src 'none'",
		fmt.Sprintf("script-src 'nonce-%s'", nonceString),
		// fmt.Sprintf("style-src-elem 'nonce-%s'", nonceString),
		"style-src-elem 'self' 'unsafe-inline'", // TODO: Use nonce when Vite fixes csp style
		"img-src 'self'",
		"manifest-src 'self'",
		"font-src 'self'",
		"connect-src 'self'",
	}

	// make csp header value
	csp := ""
	for _, s := range contentSecurityPolicy {
		csp += s + "; "
	}

	// set header (overwrites any other CSP header)
	w.Header().Set("Content-Security-Policy", csp)
}

// frontendFileHandler provides a handler for the frontend files
func (app *Application) frontendFileHandler(w http.ResponseWriter, r *http.Request) error {
	// remove the frontend URL root path (it is not used for the file path where frontend
	// is stored)
	fPath := strings.TrimPrefix(r.URL.Path, frontendUrlPath)

	// check file extension. if there is no extension, this is a path. always return index.html
	// for any path. react router will handle routing of the path from there.
	fExt := filepath.Ext(fPath)
	if fExt == "" {
		fPath = "/index.html"
		fExt = ".html"
	}

	// app.logger.Debugf("serving frontend file: %s -> %s (fext = %s)", r.URL.Path, fPath, fExt)

	// open requested file
	f, err := os.Open(frontendBuildDir + "/" + fPath)
	if err != nil {
		app.logger.Debugf("cannot find frontend file %s", fPath)
		return output.ErrNotFound
	}
	defer f.Close()

	// get file info
	fInfo, err := f.Stat()
	if err != nil {
		app.logger.Errorf("could not get file info for frontend file %s", fPath)
		return output.ErrInternal
	}

	// if fExt is of an approved type, generate a nonce, do nonce injection, and set the CSP
	if fExt == ".html" {
		// generate nonce
		nonce, err := randomness.GenerateFrontendNonce()
		if err != nil {
			app.logger.Errorf("failed to generate nonce for frontend (%s)", err)
			return output.ErrInternal
		}

		// set CSP
		setContentSecurityPolicy(w, nonce)

		// read in file to serve
		fBytes := make([]byte, fInfo.Size())
		_, err = f.Read(fBytes)
		if err != nil {
			app.logger.Errorf("could not read frontend file %s into buffer for nonce injection", fPath)
			return output.ErrInternal
		}

		// set nonce placeholders to the actual nonce value
		fBytes = bytes.ReplaceAll(fBytes, noncePlaceholder, nonce)

		// set CSP, serve modified file, and return (modtime is now since nonce is always modified)
		http.ServeContent(w, r, fInfo.Name(), time.Now(), bytes.NewReader(fBytes))
		return nil
	}

	// serve file as-is if no nonce specified
	http.ServeContent(w, r, fInfo.Name(), fInfo.ModTime(), f)
	return nil
}

// redirectToFrontendHandler is a handler that redirects to the frontend app
func redirectToFrontendHandler(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, frontendUrlPath, http.StatusPermanentRedirect)
	return nil
}

// setFrontendEnv creates the env.js file in the frontend build. This is used
// to set variables at server run time
func setFrontendEnv(frontendShowDebugInfo *bool) error {
	// remove any old environment
	_ = os.Remove(frontendEnvFile)

	// show debug info if set
	showDebugInfo := false
	if frontendShowDebugInfo != nil && *frontendShowDebugInfo {
		showDebugInfo = true
	}

	// content of new environment file
	// api and & app on same server, so use path for api url
	envFileContent := `
	window.env = {
		API_URL: '` + apiUrlPath + `',
		SHOW_DEBUG_INFO: ` + strconv.FormatBool(showDebugInfo) + `
	};
	`

	file, err := os.Create(frontendEnvFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.Write([]byte(envFileContent))
	if err != nil {
		return err
	}

	return nil
}
