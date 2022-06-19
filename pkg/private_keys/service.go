package private_keys

import (
	"log"
)

// App interface is for connecting to the main app
type App interface {
	GetKeyStorage() Storage
	GetLogger() *log.Logger
}

// Storage interface for storage functions
type Storage interface {
	GetAllKeys() ([]Key, error)
	GetOneKey(int) (Key, error)
	PutNameDescKey(NameDescPayload) error
	PostNewKey(NewPayload) error
	DeleteKey(int) error
}

// Keys service struct
type Service struct {
	storage Storage
	logger  *log.Logger
}

func NewService(app App) *Service {
	service := new(Service)

	service.storage = app.GetKeyStorage()
	service.logger = app.GetLogger()

	return service
}
