package app

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// RunLeGoAPI starts the application
func RunLeGoAPI() {
	// parse config file
	cfg := readConfigFile()

	// create the app
	app, err := Create(cfg)
	if err != nil {
		log.Panicf("panic: failed to configure app: %s", err)
	}
	defer app.CloseStorage()

	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if *cfg.DevMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", *cfg.Hostname, *cfg.Port),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch webserver
	app.GetLogger().Infof("starting lego-certhub on %s:%d", *cfg.Hostname, *cfg.Port)
	err = srv.ListenAndServe()
	if err != nil {
		app.GetLogger().Panicf("panic: failed to start http server: %s", err)
	}
}
