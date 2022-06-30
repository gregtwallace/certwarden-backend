package private_keys

import (
	"errors"
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
	GetOneKeyById(int) (Key, error)
	GetOneKeyByName(string) (Key, error)
	PutNameDescKey(NameDescPayload) error
	PostNewKey(NewPayload) error
	DeleteKey(int) error

	GetAvailableKeys() ([]Key, error)
}

// Keys service struct
type Service struct {
	logger  *log.Logger
	storage Storage
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errors.New("private_keys: newservice requires valid logger")
	}

	// storage
	service.storage = app.GetKeyStorage()
	if service.storage == nil {
		return nil, errors.New("private_keys: newservice requires valid storage")
	}

	return service, nil
}
