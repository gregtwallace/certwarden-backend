package acme_accounts

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/private_keys"
	"log"
)

// App interface is for connecting to the main app
type App interface {
	GetAccountStorage() Storage
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetLogger() *log.Logger
}

// Storage interface for storage functions
type Storage interface {
	// keys
	GetAvailableKeys() ([]private_keys.Key, error)

	// accounts
	GetAllAccounts() ([]Account, error)
	GetOneAccountById(int) (Account, error)
	GetOneAccountByName(string) (Account, error)

	PostNewAccount(NewPayload) (int, error)

	PutNameDescAccount(NameDescPayload) error
	PutLEAccountResponse(id int, response acme.AcmeAccountResponse) error

	DeleteAccount(int) error
}

// Accounts service struct
type Service struct {
	logger      *log.Logger
	storage     Storage
	acmeProd    *acme.Service
	acmeStaging *acme.Service
}

// NewService creates a new acme_accounts service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errors.New("acme_accounts: newservice requires valid logger")
	}

	// storage
	service.storage = app.GetAccountStorage()
	if service.storage == nil {
		return nil, errors.New("acme_accounts: newservice requires valid storage")
	}

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errors.New("acme_accounts: newservice requires valid acme prod service")
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errors.New("acme_accounts: newservice requires valid acme staging service")
	}

	return service, nil
}
