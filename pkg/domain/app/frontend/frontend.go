package frontend

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const buildDir = "./frontend_build"
const envFile = buildDir + "/env.js"

// runFrontend launches a web server to host the frontend
func (service *Service) run() error {
	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if service.devMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	// http server config
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *service.hostname, *service.config.HttpPort),
		Handler:      frontendRoutes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// configure and launch https if app succesfully got a cert
	if service.httpsCert != nil {
		// make tls config
		tlsConf, err := service.tlsConf()
		if err != nil {
			return err
		}

		// https server config
		srv.Addr = fmt.Sprintf("%s:%d", *service.hostname, *service.config.HttpsPort)
		srv.TLSConfig = tlsConf

		// prepare frontend env file
		err = setFrontendEnv(fmt.Sprintf("https://%s:%d", *service.hostname, *service.apiPort))
		if err != nil {
			return err
		}

		// launch https
		service.logger.Infof("starting lego-certhub frontend (https) on %s", srv.Addr)
		go func() {
			service.logger.Panic(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		// if https failed, launch localhost only http server
		// prepare frontend env file
		setFrontendEnv(fmt.Sprintf("http://%s:%d", *service.hostname, *service.apiPort))

		// launch http
		service.logger.Warnf("starting insecure lego-certhub frontend (http) on %s", srv.Addr)
		go func() {
			service.logger.Panic(srv.ListenAndServe())
		}()
	}

	return nil
}

// setFrontendEnv creates the env.js file in the frontend build. This is used
// to set variables at server run time
func setFrontendEnv(apiUrl string) error {
	// remove any old environment
	_ = os.Remove(envFile)

	// content of new environment file
	envFileContent := `
	window.env = {
		API_URL: '` + apiUrl + `',
	};
	`

	file, err := os.Create(envFile)
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
