package acme_accounts

import (
	"legocerthub-backend/pkg/private_keys"
	"legocerthub-backend/pkg/utils/acme_utils"
	"log"
)

// App interface is for connecting to the main app
type App interface {
	GetAccountStorage() Storage
	GetProdDir() *acme_utils.AcmeDirectory
	GetStagingDir() *acme_utils.AcmeDirectory
	GetLogger() *log.Logger
}

// Storage interface for storage functions
type Storage interface {
	// keys
	GetAvailableKeys() ([]private_keys.Key, error)
	// accounts
	GetAllAccounts() ([]Account, error)
	PostNewAccount(AccountPayload) (int, error)
	GetOneAccount(int) (Account, error)
	PutExistingAccount(AccountPayload) error

	// specific things
	GetAccountKid(int) (string, error)
	PutNameDescAccount(NameDescPayload) error
	GetAccountPem(int) (string, error)
	PutLEAccountInfo(id int, response acme_utils.AcmeAccountResponse) error
}

// Keys service struct
type Service struct {
	storage        Storage
	acmeProdDir    *acme_utils.AcmeDirectory
	acmeStagingDir *acme_utils.AcmeDirectory
	logger         *log.Logger
}

func NewService(app App) *Service {
	service := new(Service)

	service.storage = app.GetAccountStorage()

	service.acmeProdDir = app.GetProdDir()
	service.acmeStagingDir = app.GetStagingDir()

	service.logger = app.GetLogger()

	return service
}
