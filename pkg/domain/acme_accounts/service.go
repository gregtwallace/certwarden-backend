package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary account service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAccountStorage() Storage
	GetKeysService() *private_keys.Service
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
}

// Storage interface for storage functions
type Storage interface {
	GetAllAccounts(q pagination_sort.Query) ([]Account, int, error)
	GetOneAccountById(id int) (Account, error)
	GetOneAccountByName(name string) (Account, error)

	PostNewAccount(NewPayload) (id int, err error)

	PutNameDescAccount(NameDescPayload) error
	PutAcmeAccountResponse(response AcmeAccount) error
	PutNewAccountKey(payload RolloverKeyPayload) error

	DeleteAccount(int) error

	AccountHasCerts(accountId int) (inUse bool)

	GetOneKeyById(id int) (private_keys.Key, error)
}

// Accounts service struct
type Service struct {
	logger      *zap.SugaredLogger
	output      *output.Service
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

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
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
