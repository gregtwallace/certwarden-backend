package app

import (
	"net/http"
	"os"
	"strconv"
	"strings"

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
	r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, frontendUrlPath)

	// to handle React Router, redirect any non-existent paths to the index
	// page which will then handle Route

	// only potentially modify non '/' requests that don't contain a period in
	// the final segment (i.e. only modify requests for paths, specific file
	// names should not be re-written)
	pathParts := strings.Split(r.URL.Path, "/")
	lastPart := pathParts[len(pathParts)-1]

	// if the request is not for a specific file, modify the request
	// to return root and React Router will handle the Route (path)
	if !strings.Contains(lastPart, ".") {
		r.URL.Path = "/"
		r.URL.RawPath = "/"
	}

	// serve request from file server
	fs.ServeHTTP(w, r)

	return nil
}
