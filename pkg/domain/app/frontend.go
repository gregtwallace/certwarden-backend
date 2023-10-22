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

	"go.uber.org/zap/zapcore"
)

const frontendBuildDir = "./frontend_build"
const frontendEnvFile = frontendBuildDir + "/env.js"

// setFrontendEnv creates the env.js file in the frontend build. This is used
// to set variables at server run time
func (app *Application) setFrontendEnv() error {
	// remove any old environment
	_ = os.Remove(frontendEnvFile)

	// content of new environment file
	// api and & app on same server, so use path for api url
	envFileContent := `
	window.env = {
		API_URL: '` + apiUrlPath + `',
		SHOW_DEBUG_INFO: ` + strconv.FormatBool(app.logger.Level() == zapcore.DebugLevel) + `
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

// setFrontendSecurityHeaders sets headers that are specific to serving the frontend
// they work to mitigate things like click jacking. If the file being served is html
// the CSP nonce is also returned by this function (for use in rewriting the html file)
func setFrontendSecurityHeaders(w http.ResponseWriter, isHtmlFile bool) (nonce []byte, err error) {
	// header: CSP ("Content-Security-Policy")
	if isHtmlFile {
		nonce, err = randomness.GenerateFrontendNonce()
		if err != nil {
			return nil, fmt.Errorf("failed to generate nonce for frontend (%s)", err)
		}

		// actual security policy
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
		csp := ""
		for _, s := range contentSecurityPolicy {
			csp += s + "; "
		}

		w.Header().Set("Content-Security-Policy", csp)

	} else {
		// if not html, do not allow any sources
		w.Header().Set("Content-Security-Policy", "default-src 'none'; ")
	}

	// header: no MIME type sniffing (strict MIME) ("X-Content-Type-Options")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// header: do NOT allow frames ("X-Frame-Options")
	w.Header().Set("X-Frame-Options", "deny")

	return nonce, nil
}

// frontendHandler provides a handler for the frontend
func (app *Application) frontendHandler(w http.ResponseWriter, r *http.Request) error {
	// remove the frontend URL root path (it is not used for the file path where frontend
	// is stored)
	fPath := strings.TrimPrefix(r.URL.Path, frontendUrlPath)

	// check file extension. if there is no extension, this is a path. always return index.html
	// for any path. react router will handle routing of the path from there.
	fExt := filepath.Ext(fPath)
	if fExt == "" {
		fPath = "/index.html"
	}
	fExt = filepath.Ext(fPath)

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

	// is this file .html?
	isHtmlFile := fExt == ".html"

	// set security headers
	nonce, err := setFrontendSecurityHeaders(w, isHtmlFile)
	if err != nil {
		app.logger.Error(err)
		return output.ErrInternal
	}

	// if is html file, look for nonce placeholders, update them, and then serve
	// the modified buffer
	if isHtmlFile {
		// read in
		fBytes := make([]byte, fInfo.Size())
		_, err = f.Read(fBytes)
		if err != nil {
			app.logger.Errorf("could not read frontend .html file %s into buffer", fPath)
			return output.ErrInternal
		}

		// set nonce placeholders to the actual nonce value
		noncePlaceholder := []byte("{SERVER-CSP-NONCE}")
		fBytes = bytes.ReplaceAll(fBytes, noncePlaceholder, nonce)

		// set CSP, serve modified file, and return (modtime is now since nonce is always modified)
		http.ServeContent(w, r, fInfo.Name(), time.Now(), bytes.NewReader(fBytes))
		return nil
	}

	// serve file as-is if not .html
	http.ServeContent(w, r, fInfo.Name(), fInfo.ModTime(), f)
	return nil
}
