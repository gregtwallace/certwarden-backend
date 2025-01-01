package auth

import (
	"certwarden-backend/pkg/domain/app/auth/session_manager"
	"certwarden-backend/pkg/output"
	"context"
	"errors"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary auth service component is missing")

// constant for bcrypt cost value
const BcryptCost = 12

// App interface is for connecting to the main app
type App interface {
	IsHttps() bool
	CORSPermittedCrossOrigins() []string
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
	logger         *zap.SugaredLogger
	output         *output.Service
	storage        Storage
	sessionManager *session_manager.SessionManager
}

// NewService creates a new users service
func NewService(app App) (*Service, error) {
	service := new(Service)

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

	// create session manager
	service.sessionManager = session_manager.NewSessionManager(app.IsHttps(), len(app.CORSPermittedCrossOrigins()) > 0, service.logger)
	// start cleaner
	// service.startCleanerService(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	return service, nil
}

// make ValidateAuthHeader available to App
func (service *Service) ValidateAuthHeader(r *http.Request, w http.ResponseWriter, logTaskName string) (username string, _ error) {
	return service.sessionManager.ValidateAuthHeader(r, w, logTaskName)
}
