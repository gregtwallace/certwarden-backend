package main

import (
	"flag"
	"fmt"
	"legocerthub-backend/pkg/app"
	"legocerthub-backend/pkg/storage/sqlite"

	"log"
	"net/http"
	"net/url"
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

	// sqlite options - see: https://github.com/mattn/go-sqlite3#connection-string
	// should enforce foreign key constraints
	cfg.Db.Options = url.Values{}
	cfg.Db.Options.Add("_fk", "true")

	// open database connection
	storage, err := sqlite.OpenDB(cfg.Db.Dsn, cfg.Db.Options)
	if err != nil {
		logger.Fatalln(err)
	}
	defer storage.Db.Close()

	// TODO: setup nonce management

	// configure the Application
	app := &app.Application{
		Config:  cfg,
		Logger:  logger,
		Storage: storage,
	}

	// initialize directory structs (avoid nil pointers)
	app.InitDirectories()
	// populate directory structs on the app
	err = app.UpdateAllDirectories()
	if err != nil {
		logger.Fatalln(err)
	}
	// & start background process to check for updates periodically
	go app.BackgroundDirManagement()

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
