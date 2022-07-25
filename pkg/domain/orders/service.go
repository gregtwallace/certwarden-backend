package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary orders service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetOrderStorage() Storage
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetAuthzService() *authorizations.Service
}

// Storage interface for storage functions
type Storage interface {
	GetOneCertById(id int, withAcctPem bool) (cert certificates.Certificate, err error)

	// orders
	PostNewOrder(newOrder Order) (newId int, err error)
}

// Keys service struct
type Service struct {
	logger         *zap.SugaredLogger
	output         *output.Service
	storage        Storage
	acmeProd       *acme.Service
	acmeStaging    *acme.Service
	authorizations *authorizations.Service
}

// NewService creates a new private_key service
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
	service.storage = app.GetOrderStorage()
	if service.storage == nil {
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

	// authorization service
	service.authorizations = app.GetAuthzService()
	if service.authorizations == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
