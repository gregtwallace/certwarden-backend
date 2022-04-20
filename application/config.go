package application

import (
	"legocerthub-backend/database"
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
