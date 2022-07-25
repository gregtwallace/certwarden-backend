package authorizations

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/acme/challenges/http01"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary authorizations service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetHttp01Service() *http01.Service
}

// Storage interface for storage functions
type Storage interface {
}

// service struct
type Service struct {
	logger      *zap.SugaredLogger
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	http01      *http01.Service
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// http-01 challenge service
	service.http01 = app.GetHttp01Service()
	if service.http01 == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
