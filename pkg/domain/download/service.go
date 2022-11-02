package download

import (
	"errors"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary download service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
	IsHttps() bool
	GetOutputter() *output.Service
	GetDownloadStorage() Storage
}

// Storage interface for storage functions
type Storage interface {
	GetOneKeyByName(name string) (private_keys.Key, error)

	GetOneCertByName(name string) (cert certificates.Certificate, err error)
	GetCertPemById(certId int) (pem string, err error)
}

// Keys service struct
type Service struct {
	devMode bool
	logger  *zap.SugaredLogger
	https   bool
	output  *output.Service
	storage Storage
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// devMode
	service.devMode = app.GetDevMode()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// running as https?
	service.https = app.IsHttps()

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetDownloadStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
