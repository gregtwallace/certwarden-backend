package acme_accounts

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"certwarden-backend/pkg/domain/private_keys"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary account service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAccountStorage() Storage
	GetKeysService() *private_keys.Service
	GetAcmeServerService() *acme_servers.Service
}

// Storage interface for storage functions
type Storage interface {
	GetAllAcmeAccounts(q pagination_sort.Query) ([]Account, int, error)
	GetOneAcmeAccountById(id int) (Account, error)
	GetOneAcmeAccountByName(name string) (Account, error)

	PostNewAcmeAccount(payload NewPayload) (newAcct Account, err error)

	PutAcmeAccountUpdate(payload *UpdatePayload) (updatedAcct Account, err error)
	PutAcmeAccountNewKey(payload RolloverKeyPayload) (updatedAcct Account, err error)

	DeleteAcmeAccount(int) error

	AcmeAccountInUse(accountId int) (inUse bool, err error)

	GetOneKeyById(id int) (private_keys.Key, error)
}

// Accounts service struct
type Service struct {
	logger            *zap.SugaredLogger
	output            *output.Service
	storage           Storage
	keys              *private_keys.Service
	acmeServerService *acme_servers.Service
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
	service.acmeServerService = app.GetAcmeServerService()
	if service.acmeServerService == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
