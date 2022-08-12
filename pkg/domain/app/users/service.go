package users

import (
	"errors"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary users service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetUsersStorage() Storage
}

type Storage interface {
	GetOneUserByName(username string) (User, error)
}

// Keys service struct
type Service struct {
	logger  *zap.SugaredLogger
	output  *output.Service
	storage Storage
}

// NewService creates a new (local LeGo) users service
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
	service.storage = app.GetUsersStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
