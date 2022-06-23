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
	var err error

	// configure the app
	app, err := app.CreateAndConfigure()
	if err != nil {
		log.Fatalln(err)
	}

	// parse command line for web server config
	var webCfg webConfig

	flag.StringVar(&webCfg.host, "host", "localhost", "hostname to listen on")
	flag.IntVar(&webCfg.port, "port", 4050, "port number to listen on")
	flag.Parse()

	// configure webserver
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", webCfg.host, webCfg.port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// launch webserver
	app.GetLogger().Println("Starting server on host", webCfg.host, "port", webCfg.port)
	err = srv.ListenAndServe()
	if err != nil {
		app.GetLogger().Println(err)
	}
}
