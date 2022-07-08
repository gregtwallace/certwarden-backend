package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary account service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetAccountStorage() Storage
	GetKeysService() *private_keys.Service
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetLogger() *zap.SugaredLogger
}

// Storage interface for storage functions
type Storage interface {
	GetAllAccounts() ([]Account, error)
	GetOneAccountById(int) (Account, error)
	GetOneAccountByName(string) (Account, error)

	PostNewAccount(NewPayload) (id int, err error)

	PutNameDescAccount(NameDescPayload) error
	PutLEAccountResponse(id int, response acme.AcmeAccountResponse) error

	DeleteAccount(int) error
}

// Accounts service struct
type Service struct {
	logger      *zap.SugaredLogger
	storage     Storage
	keys        *private_keys.Service
	acmeProd    *acme.Service
	acmeStaging *acme.Service
}

// NewService creates a new acme_accounts service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetAccountStorage()
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

	return service, nil
}
