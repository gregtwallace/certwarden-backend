package dns01cloudflare

import (
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	logger *zap.SugaredLogger
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
