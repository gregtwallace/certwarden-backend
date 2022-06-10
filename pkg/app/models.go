package app

import (
	"legocerthub-backend/pkg/storage/sqlite"
	"legocerthub-backend/pkg/utils/acme_utils"
	"log"
)

const version = "0.0.1"

type Config struct {
	Host string
	Port int
	Env  string
}

type AppAcme struct {
	ProdDir    *acme_utils.AcmeDirectory
	StagingDir *acme_utils.AcmeDirectory
}

type Application struct {
	Config  Config
	Logger  *log.Logger
	Storage *sqlite.Storage
	Acme    AppAcme
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
