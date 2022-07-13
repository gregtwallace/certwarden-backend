package output

import (
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary output service component is missing")

type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
}

type Service struct {
	devMode bool
	logger  *zap.SugaredLogger
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// devmode?
	service.devMode = app.GetDevMode()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
