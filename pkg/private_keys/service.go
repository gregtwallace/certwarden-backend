package private_keys

import (
	"log"
	"net/http"
)

// Keys service to provide handlers
type Service interface {
	GetAllKeys(http.ResponseWriter, *http.Request)
	GetOneKey(http.ResponseWriter, *http.Request)
	GetNewKeyOptions(http.ResponseWriter, *http.Request)
	PutOneKey(http.ResponseWriter, *http.Request)
	PostNewKey(http.ResponseWriter, *http.Request)
	DeleteKey(http.ResponseWriter, *http.Request)
}

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
type service struct {
	storage Storage
	logger  *log.Logger
}

func NewService(app App) Service {
	service := new(service)

	service.storage = app.GetStorage()
	service.logger = app.GetLogger()

	return service
}
