package certificates

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/http01"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary cert service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetCertificatesStorage() Storage
	GetKeysService() *private_keys.Service
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetAcctsService() *acme_accounts.Service
	GetHttp01Service() *http01.Service
}

// Storage interface for storage functions
type Storage interface {
	GetAllCerts() (certs []Certificate, err error)
	GetOneCertById(id int, withAcctPem bool) (cert Certificate, err error)
	GetOneCertByName(name string, withAcctPem bool) (cert Certificate, err error)

	PostNewCert(payload NewPayload) (id int, err error)
	PostNewOrder(cert Certificate, response acme.OrderResponse) (newId int, err error)

	PutDetailsCert(payload DetailsUpdatePayload) (err error)

	DeleteCert(id int) (err error)

	GetOneAccountById(id int, withPem bool) (acme_accounts.Account, error)
}

// Keys service struct
type Service struct {
	logger      *zap.SugaredLogger
	output      *output.Service
	storage     Storage
	keys        *private_keys.Service
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	accounts    *acme_accounts.Service
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

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// account services
	service.accounts = app.GetAcctsService()
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
