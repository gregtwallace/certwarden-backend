package app

import (
	"context"
	"database/sql"
	"legocerthub-backend/database"
	"log"
	"time"
)

const version = "0.0.1"

type Config struct {
	Host string
	Port int
	Env  string
	Db   struct {
		Dsn string
	}
}

type AppStatus struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type Application struct {
	Config   Config
	Logger   *log.Logger
	Database *database.DBWrap
}

// function opens connection to the sqlite database
//   this will also cause the file to be created, if it does not exist
func OpenDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
