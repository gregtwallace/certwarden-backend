package certificates

import (
	"errors"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary key service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetCertificatesStorage() Storage
}

// Storage interface for storage functions
type Storage interface {
	GetAllCertificates() ([]Certificate, error)
}

// Keys service struct
type Service struct {
	logger  *zap.SugaredLogger
	output  *output.Service
	storage Storage
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
	service.storage = app.GetCertificatesStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
