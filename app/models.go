package app

import (
	"database/sql"
	"log"
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

type Application struct {
	Config Config
	Logger *log.Logger
	DB     *sql.DB
}

type appStatus struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}
