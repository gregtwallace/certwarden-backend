package frontend

import (
	"fmt"
	"net/http"
	"time"
)

const buildDir = "./frontend_build"

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
		Handler:      service.frontendRoutes(),
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
