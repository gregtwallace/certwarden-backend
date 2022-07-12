package main

import (
	"flag"
	"fmt"
	"legocerthub-backend/pkg/domain/app"
	"log"

	"net/http"
	"time"
)

// http server config options
type webConfig struct {
	host string
	port int
}

func main() {
	// parse command line for config options
	var webCfg webConfig
	var devMode bool

	flag.StringVar(&webCfg.host, "host", "localhost", "hostname to listen on")
	flag.IntVar(&webCfg.port, "port", 4050, "port number to listen on")
	// TODO: change default to false
	flag.BoolVar(&devMode, "development", true, "run the server in development mode")
	flag.Parse()

	// configure the app
	app, err := app.CreateAndConfigure(devMode)
	if err != nil {
		log.Fatalln(err)
	}
	defer app.CloseStorage()

	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if devMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", webCfg.host, webCfg.port),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch webserver
	app.GetLogger().Infof("starting lego-certhub on %s:%d", webCfg.host, webCfg.port)
	err = srv.ListenAndServe()
	if err != nil {
		app.GetLogger().Panicf("failed to start server %s", err)
	}
}
