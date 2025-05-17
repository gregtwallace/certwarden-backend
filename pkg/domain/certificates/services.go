package certificates

import (
	"certwarden-backend/pkg/domain/acme_accounts"
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary cert service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetCertificatesStorage() Storage
	GetAcmeServerService() *acme_servers.Service
	GetKeysService() *private_keys.Service
	GetAcctsService() *acme_accounts.Service
}

// Storage interface for storage functions
type Storage interface {
	GetAllCerts(q pagination_sort.Query) (certs []Certificate, totalRowCount int, err error)
	GetOneCertById(id int) (cert Certificate, err error)
	GetOneCertByName(name string) (cert Certificate, err error)

	PostNewCert(payload NewPayload) (Certificate, error)

	PutDetailsCert(payload DetailsUpdatePayload) (Certificate, error)
	PutCertApiKey(certId int, apiKey string, updateTimeUnix int) (err error)
	PutCertNewApiKey(certId int, newApiKey string, updateTimeUnix int) (err error)
	PutCertClientKey(certId int, newClientKeyB64 string, updateTimeUnix int) (err error)

	DeleteCert(id int) (err error)

	PostNewKey(private_keys.NewPayload) (private_keys.Key, error)
}

// Keys service struct
type Service struct {
	logger            *zap.SugaredLogger
	output            *output.Service
	storage           Storage
	acmeServerService *acme_servers.Service
	keys              *private_keys.Service
	accounts          *acme_accounts.Service
}

// NewService creates a new service
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

	// acme services
	service.acmeServerService = app.GetAcmeServerService()
	if service.acmeServerService == nil {
		return nil, errServiceComponent
	}

	// key service
	service.keys = app.GetKeysService()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	// account services
	service.accounts = app.GetAcctsService()
	if service.accounts == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
