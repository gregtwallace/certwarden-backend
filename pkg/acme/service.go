package acme

import (
	"errors"
	"log"
)

// App interface is for connecting to the main app
type App interface {
	//GetAccountStorage() Storage
	GetLogger() *log.Logger
}

// Acme service struct
type Service struct {
	logger *log.Logger
	dirUri string
	dir    *acmeDirectory
	// nonce manager
}

// NewService creates a new acme service based on a directory uri
func NewService(app App, dirUri string) (*Service, error) {
	service := new(Service)
	var err error

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errors.New("acme: newservice requires valid storage")
	}

	// acme directory
	service.dirUri = dirUri
	service.dir = new(acmeDirectory)

	// initial population
	err = service.fetchAcmeDirectory()
	if err != nil {
		return nil, err
	}

	// start go routine to check for periodic updates
	service.backgroundDirManager()

	return service, nil
}
