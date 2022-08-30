package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/providers/http01internal"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary challenges service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetDevMode() bool
}

// interface for any provider service
type providerService interface {
	Provision(resourceName string, resourceContent string) (err error)
	Deprovision(resourceName string) (err error)
}

// service struct
type Service struct {
	logger             *zap.SugaredLogger
	acmeProd           *acme.Service
	acmeStaging        *acme.Service
	challengeProviders challengeProviders
}

// service challenge providers
type challengeProviders struct {
	http01Internal providerService
}

// NewService creates a new service
func NewService(app App) (service *Service, err error) {
	service = new(Service)

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

	// challenge providers

	// http-01 internal challenge server
	// TODO: Don't hardcode port
	service.challengeProviders.http01Internal, err = http01internal.NewService(app, 4060)
	if err != nil {
		return nil, err
	}

	return service, nil
}
