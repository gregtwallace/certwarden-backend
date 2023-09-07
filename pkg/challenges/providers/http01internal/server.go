package http01internal

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	service.logger.Infof("attempting to start http-01 challenge server on %s.", servAddr)
	if service.port != 80 {
		service.logger.Warnf("http-01 challenge server is not configured on port 80; internet "+
			"facing port 80 must be proxied to port %d to function.", service.port)
	}

	// create listener for web server
	ln, err := net.Listen("tcp", servAddr)
	if err != nil {
		service.logger.Error(fmt.Errorf("failed to start http-01 challenge server, cannot bind to %s (%s)", servAddr, err))
		return err
	}

	// start server
	service.shutdownWaitgroup.Add(1)
	go func() {
		defer func() { _ = ln.Close }()
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			service.logger.Errorf("http01internal server returned error (%s)", err)
		}
		service.logger.Infof("http-01 challenge server (%s) shutdown complete", servAddr)
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
			service.logger.Errorf("error shutting down http-01 challenge server %s (%s)", servAddr, err)
		}

		// send shutdown result to err chan
		service.stopErrChan <- err
	}()

	return nil
}
