package http01

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
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch webserver
	service.logger.Infof("starting http-01 challenge server on %s.", servAddr)
	service.logger.Warn("http-01 challenge server is not running on port 80; requests must be proxied to part 80 or they will not pass")
	go func() {
		service.logger.Panic(srv.ListenAndServe())
	}()

	return nil
}
