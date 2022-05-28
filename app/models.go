package app

import (
	"database/sql"
	"legocerthub-backend/utils/acme_utils"
	"log"
	"net/url"
	"time"
)

const version = "0.0.1"

type Config struct {
	Host string
	Port int
	Env  string
	Db   struct {
		Dsn     string
		Options url.Values
	}
}

type AppDb struct {
	Database *sql.DB
	Timeout  time.Duration
}

type AppAcme struct {
	ProdDir    *acme_utils.AcmeDirectory
	StagingDir *acme_utils.AcmeDirectory
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
