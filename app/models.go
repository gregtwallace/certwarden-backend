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
	Database *sql.DB
	Timeout  time.Duration
}

type Application struct {
	Config Config
	Logger *log.Logger
	DB     AppDb
	Acme   AppAcme
}

type appStatusDirectories struct {
	Production string `json:"prod"`
	Staging    string `json:"staging"`
}

type appStatus struct {
	Status          string               `json:"status"`
	Environment     string               `json:"environment"`
	Version         string               `json:"version"`
	AcmeDirectories appStatusDirectories `json:"acme_directories"`
}
