package app

import (
	"database/sql"
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

type AppDb struct {
	DB      *sql.DB
	Timeout time.Duration
}

type Application struct {
	Config Config
	Logger *log.Logger
	DB     AppDb
}

type appStatus struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}
