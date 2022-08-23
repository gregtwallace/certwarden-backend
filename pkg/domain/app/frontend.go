package app

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const buildDir = "./frontend_build"
const envFile = buildDir + "/env.js"

// runFrontend launches a web server to host the frontend
func (app *Application) runFrontend() {
	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if *app.config.DevMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	// http server config
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *app.config.Hostname, *app.config.Frontend.HttpPort),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// make file server handle for build dir
	http.Handle("/", http.FileServer(http.Dir(buildDir)))

	// configure and launch https if app succesfully got a cert
	if app.appCert != nil {
		// make tls config
		tlsConf, err := app.TlsConf()
		if err != nil {
			app.logger.Panicf("tls config problem: %s", err)
			return
		}

		// https server config
		srv.Addr = fmt.Sprintf("%s:%d", *app.config.Hostname, *app.config.Frontend.HttpsPort)
		srv.TLSConfig = tlsConf

		// prepare frontend env file
		err = setFrontendEnv(fmt.Sprintf("https://%s", srv.Addr))
		if err != nil {
			app.logger.Panicf("error setting frontend environment: %s", err)
			return
		}

		// launch https
		app.logger.Infof("starting lego-certhub frontend (https) on %s", srv.Addr)
		app.logger.Panic(srv.ListenAndServeTLS("", ""))
	} else {
		// if https failed, launch localhost only http server
		// prepare frontend env file
		setFrontendEnv(fmt.Sprintf("http://%s", srv.Addr))

		// launch http
		app.logger.Warnf("starting insecure lego-certhub frontend (http) on %s", srv.Addr)
		app.logger.Panic(srv.ListenAndServe())
	}
}

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
