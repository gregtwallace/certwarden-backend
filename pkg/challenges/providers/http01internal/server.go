package http01internal

import (
	"fmt"
	"net/http"
	"time"
)

func (service *Service) startServer(port int) (err error) {
	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if service.devMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	// TODO: modify to allow specifying specific interface addresses
	hostName := ""

	servAddr := fmt.Sprintf("%s:%d", hostName, port)
	srv := &http.Server{
		Addr:         servAddr,
		Handler:      service.routes(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// no need to keep these connections alive
	srv.SetKeepAlivesEnabled(false)

	// launch webserver
	service.logger.Infof("starting http-01 challenge server on %s.", servAddr)
	if port != 80 {
		service.logger.Warnf("http-01 challenge server is not running on port 80; internet "+
			"facing port 80 must be proxied to port %d to function.", port)
	}

	go func() {
		service.logger.Panic(srv.ListenAndServe())
	}()

	return nil
}
