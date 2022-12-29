package http01internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func (service *Service) startServer(port int, ctx context.Context, wg *sync.WaitGroup) (err error) {
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
	wg.Add(1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			service.logger.Panic(err)
		}
		service.logger.Info("http-01 challenge server shutdown complete")
		wg.Done()
	}()

	// monitor shutdown context
	go func() {
		<-ctx.Done()

		maxShutdownTime := 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), maxShutdownTime)
		defer cancel()

		err = srv.Shutdown(ctx)
		if err != nil {
			service.logger.Errorf("error shutting down http-01 challenge server")
		}
	}()

	return nil
}
