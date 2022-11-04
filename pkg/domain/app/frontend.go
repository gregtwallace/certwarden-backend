package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const frontendBuildDir = "./frontend_build"
const frontendEnvFile = frontendBuildDir + "/env.js"
const frontendUrlPath = "/app"

// setFrontendEnv creates the env.js file in the frontend build. This is used
// to set variables at server run time
func (app *Application) setFrontendEnv() error {
	// remove any old environment
	_ = os.Remove(frontendEnvFile)

	// content of new environment file
	envFileContent := `
	window.env = {
		API_URL: '` + *app.config.Hostname + `',
		DEV_MODE: ` + strconv.FormatBool(*app.config.DevMode) + `
	};
	`

	file, err := os.Create(frontendEnvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(envFileContent))
	if err != nil {
		return err
	}

	return nil
}

// frontendHandler provides a handler for the frontend
func (app *Application) frontendHandler(w http.ResponseWriter, r *http.Request) error {
	// file server for frontend
	fs := http.FileServer(http.Dir(frontendBuildDir))

	// remove the frontend URL root path
	r.URL.Path = strings.TrimPrefix(r.URL.Path, frontendUrlPath)

	// only potentially modify non '/' requests that dont contain a period in
	// the final segment (i.e. only modify requests for paths, specific file
	// names should still be able to 404)
	pathParts := strings.Split(r.URL.Path, "/")
	lastPart := pathParts[len(pathParts)-1]
	// check path is not / AND last part of the path does NOT contain a period (i.e. not a file)
	if r.URL.Path != "/" && !strings.Contains(lastPart, ".") {
		// check if request (as-is) exists
		fullPath := frontendBuildDir + r.URL.Path

		_, err := os.Stat(fullPath)
		if err != nil {
			// confirm error is file doesn't exist
			if !os.IsNotExist(err) {
				// if some other error, log it and return 404
				app.logger.Errorf("error serving frontend: %s", err)
				return output.ErrNotFound
			}

			// if path doesn't exist, redirect to frontend root (index)
			redirectToFrontendHandler(w, r)
			return nil
		}
	}

	// exists - serve request from file server
	fs.ServeHTTP(w, r)

	return nil

}

// redirectToFrontendHandler redirects to the root of the frontend
func redirectToFrontendHandler(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, frontendUrlPath, http.StatusTemporaryRedirect)
	return nil
}
