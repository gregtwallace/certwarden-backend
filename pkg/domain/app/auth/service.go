package auth

import (
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/randomness"
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary auth service component is missing")

// constant for bcrypt cost value
const BcryptCost = 12

// App interface is for connecting to the main app
type App interface {
	IsHttps() bool
	AllowsSomeCrossOrigin() bool
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAuthStorage() Storage
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    int
	UpdatedAt    int
}

type Storage interface {
	GetOneUserByName(username string) (User, error)
	UpdateUserPassword(username string, newPasswordHash string) (userId int, err error)
}

// Keys service struct
type Service struct {
	logger           *zap.SugaredLogger
	https            bool
	allowCrossOrigin bool
	output           *output.Service
	storage          Storage
	accessJwtSecret  []byte
	sessionJwtSecret []byte
	sessionManager   *sessionManager
}

// NewService creates a new users service
func NewService(app App) (*Service, error) {
	service := new(Service)
	var err error

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// running as https?
	service.https = app.IsHttps()

	// cross origin allowed?
	service.allowCrossOrigin = app.AllowsSomeCrossOrigin()

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

	// generate new secrets on every start. this will auto-invalidate old keys to avoid any conflicts
	// caused by losing the session states on restart
	service.accessJwtSecret, err = randomness.GenerateHexSecret()
	if err != nil {
		return nil, errServiceComponent
	}

	service.sessionJwtSecret, err = randomness.GenerateHexSecret()
	if err != nil {
		return nil, errServiceComponent
	}

	// create session manager
	service.sessionManager = newSessionManager()
	// start cleaner
	service.startCleanerService(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	return service, nil
}
