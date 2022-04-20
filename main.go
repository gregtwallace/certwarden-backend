package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "0.0.1"

type config struct {
	host string
	port int
	env  string
	db   struct {
		dsn string
	}
}

type AppStatus struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type application struct {
	config   config
	logger   *log.Logger
	database *DBWrap
}

func main() {
	var cfg config

	flag.StringVar(&cfg.host, "host", "localhost", "hostname to listen on")
	flag.IntVar(&cfg.port, "port", 4050, "port number to listen on")
	flag.StringVar(&cfg.env, "env", "dev", "application environment (dev | prod)")
	flag.StringVar(&cfg.db.dsn, "dsn", "./lego-certhub.db", "database path and filename")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	app := &application{
		config: cfg,
		logger: logger,
		database: &DBWrap{
			DB: db,
		},
	}

	// create tables in the database if they don't exist
	app.database.createDBTables()

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.host, cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Println("Starting server on host", cfg.host, "port", cfg.port)

	err = srv.ListenAndServe()
	if err != nil {
		logger.Println(err)
	}

}
