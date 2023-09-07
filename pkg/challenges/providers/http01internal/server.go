package http01internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func (service *Service) startServer() (err error) {
	// make child context for stopping server
	ctx, stopServer := context.WithCancel(service.shutdownContext)
	service.stopServerFunc = stopServer

	// err chan for stop
	service.stopErrChan = make(chan error)

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

	servAddr := fmt.Sprintf("%s:%d", hostName, service.port)
	srv := &http.Server{
		Addr:         servAddr,
		Handler:      service.routes(),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// no need to keep these connections alive
	srv.SetKeepAlivesEnabled(false)

	// launch webserver
	service.logger.Infof("starting http-01 challenge server on %s for domains %s.", servAddr, service.AvailableDomains())
	if service.port != 80 {
		service.logger.Warnf("http-01 challenge server for domains %s is not running on port 80; internet "+
			"facing port 80 must be proxied to port %d to function.", service.AvailableDomains(), service.port)
	}
	service.shutdownWaitgroup.Add(1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			service.logger.Panic(err)
		}
		service.logger.Infof("http-01 challenge server (%s) shutdown complete", service.AvailableDomains())
		service.shutdownWaitgroup.Done()
	}()

	// monitor shutdown context
	go func() {
		<-ctx.Done()

		maxShutdownTime := 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = srv.Shutdown(ctx)
		if err != nil {
			service.logger.Errorf("error shutting down http-01 challenge server %s (%s)", service.AvailableDomains(), err)
		}

		// send shutdown result to err chan
		service.stopErrChan <- err
	}()

	return nil
}
