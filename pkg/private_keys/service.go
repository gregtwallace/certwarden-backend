package private_keys

import (
	"log"
)

// App interface is for connecting to the main app
type App interface {
	GetStorage() Storage
	GetLogger() *log.Logger
}

// Storage interface for storage functions
type Storage interface {
	GetAllKeys() ([]Key, error)
	GetOneKey(int) (Key, error)
	PutExistingKey(KeyPayload) error
	PostNewKey(KeyPayload) error
	DeleteKey(int) error
}

// Keys service struct
type Service struct {
	storage Storage
	logger  *log.Logger
}

func NewService(app App) *Service {
	service := new(Service)

	service.storage = app.GetStorage()
	service.logger = app.GetLogger()

	return service
}
