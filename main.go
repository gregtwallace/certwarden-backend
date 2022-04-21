package main

import (
	"flag"
	"fmt"
	"legocerthub-backend/app"

	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// create logger
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// parse command line and set config
	var cfg app.Config

	flag.StringVar(&cfg.Host, "host", "localhost", "hostname to listen on")
	flag.IntVar(&cfg.Port, "port", 4050, "port number to listen on")
	flag.StringVar(&cfg.Env, "env", "dev", "application environment (dev | prod)")
	flag.StringVar(&cfg.Db.Dsn, "dsn", "./lego-certhub.db", "database path and filename")
	flag.Parse()

	// open database connection
	db, err := app.OpenDB(cfg)
	if err != nil {
		logger.Fatalln(err)
	}
	defer db.Close()

	// configure the Application
	app := &app.Application{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}

	// create tables in the database if they don't exist
	app.CreateDBTables()

	// configure the webserver
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// launch the webserver
	logger.Println("Starting server on host", cfg.Host, "port", cfg.Port)
	err = srv.ListenAndServe()
	if err != nil {
		logger.Println(err)
	}
}
