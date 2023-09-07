package certificates

import (
	"errors"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary cert service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
	IsHttps() bool
	GetOutputter() *output.Service
	GetCertificatesStorage() Storage
	GetKeysService() *private_keys.Service
	GetAcctsService() *acme_accounts.Service
}

// Storage interface for storage functions
type Storage interface {
	GetAllCerts(q pagination_sort.Query) (certs []Certificate, totalRowCount int, err error)
	GetOneCertById(id int) (cert Certificate, err error)
	GetOneCertByName(name string) (cert Certificate, err error)

	PostNewCert(payload NewPayload) (id int, err error)

	PutDetailsCert(payload DetailsUpdatePayload) (err error)
	PutCertApiKey(certId int, apiKey string, updateTimeUnix int) (err error)
	PutCertNewApiKey(certId int, newApiKey string, updateTimeUnix int) (err error)

	DeleteCert(id int) (err error)

	PostNewKey(private_keys.NewPayload) (keyId int, err error)
}

// Keys service struct
type Service struct {
	devMode  bool
	logger   *zap.SugaredLogger
	https    bool
	output   *output.Service
	storage  Storage
	keys     *private_keys.Service
	accounts *acme_accounts.Service
}

// NewService creates a new service
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
	service.storage = app.GetCertificatesStorage()
	if service.storage == nil {
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
