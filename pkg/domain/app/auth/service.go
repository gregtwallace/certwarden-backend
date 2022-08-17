package auth

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/randomness"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary auth service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAuthStorage() Storage
}

type Storage interface {
	GetOneUserByName(username string) (User, error)
}

// Keys service struct
type Service struct {
	devMode          bool
	logger           *zap.SugaredLogger
	output           *output.Service
	storage          Storage
	accessJwtSecret  []byte
	refreshJwtSecret []byte
	sessionManager   *sessionManager
}

// NewService creates a new (local LeGo) users service
func NewService(app App) (*Service, error) {
	service := new(Service)
	var err error

	// dev mode check
	service.devMode = app.GetDevMode()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetAuthStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	// generate new secrets on every start
	// this will auto-invalidate old keys to avoid any conflicts caused
	// by losing the session states on restart
	service.accessJwtSecret, err = randomness.GenerateHexSecret()
	if err != nil {
		return nil, errServiceComponent
	}

	service.refreshJwtSecret, err = randomness.GenerateHexSecret()
	if err != nil {
		return nil, errServiceComponent
	}

	// create session manager
	service.sessionManager = newSessionManager(service.devMode)
	// start cleaner
	service.sessionManager.cleaner()

	return service, nil
}
