package app

import (
	"bytes"
	"fmt"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
		"base-uri 'none'",
		"form-action 'none'",
		"frame-ancestors 'none'",

		// scripts
		"script-src 'self'",      // fallback csp v1
		"script-src-elem 'self'", // csp v3
		"script-src-attr 'none'", // csp v3

		// styles
		fmt.Sprintf("style-src 'self' 'nonce-%s' 'unsafe-inline'", nonceString),      // fallback csp v1, unsafe-inline is for browsers that don't support nonce
		fmt.Sprintf("style-src-elem 'self' 'nonce-%s' 'unsafe-inline'", nonceString), // csp v3, unsafe-inline is for browsers that don't support nonce
		"style-src-attr 'none'", // csp v3

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

	// TODO: REMOVE WHEN PROPER NONCE SUPPORT IN VITE!
	// this modifies the code of the relevant module to be able to inject the nonce from meta tag
	// if emotion_sheet-*.js file, modify code to enable getting nonce from index.html's meta tag
	if strings.HasPrefix(fPath, "/assets/emotion_sheet-") && fExt == ".js" {
		// read in file to serve
		fBytes := make([]byte, fInfo.Size())
		_, err = f.Read(fBytes)
		if err != nil {
			app.logger.Errorf("could not read frontend file %s into buffer for nonce injection", fPath)
			return output.ErrInternal
		}

		// replace offending line of code to make it get the nonce from meta nonce
		// capture 1st, 2nd, and 3rd variable name
		// regex should cover all cases of the code, even if formatted or var names change
		re := regexp.MustCompile(`,\s*([A-Za-z0-9]+)\.nonce.*!==.*void 0.*&&.*([A-Za-z0-9]+)\.setAttribute\(["']nonce["'],.*([A-Za-z0-9]+)\.nonce\),`)
		// use 2nd variable name in new string
		fString := string(fBytes)
		fString = re.ReplaceAllString(fString, ",$2.setAttribute('nonce',document.querySelector('meta[property=\"csp-nonce\"]').nonce),")
		// orig:             ,n.nonce!==void 0&&t.setAttribute("nonce",n.nonce),
		// orig (formatted): , n.nonce !== void 0 && t.setAttribute('nonce', n.nonce),
		// modified:         ,t.setAttribute('nonce',document.querySelector('meta[property="csp-nonce"]').nonce),

		// serve modified file, and return
		http.ServeContent(w, r, fInfo.Name(), fInfo.ModTime(), strings.NewReader(fString))
		return nil
	}
	// END - TODO: REMOVE WHEN PROPER NONCE SUPPORT IN VITE

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

		// serve modified file, and return (modtime is now since nonce is always modified)
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
