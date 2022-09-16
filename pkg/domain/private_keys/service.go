package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/output"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary key service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetKeyStorage() Storage
}

// Storage interface for storage functions
type Storage interface {
	GetAllKeys() ([]Key, error)
	GetOneKeyById(id int, withPem bool) (Key, error)
	GetOneKeyByName(name string, withPem bool) (Key, error)

	PostNewKey(NewPayload) (keyId int, err error)
	PutKeyInfo(InfoPayload) error
	DeleteKey(int) error

	GetAvailableKeys() ([]Key, error)
	KeyInUse(id int) (inUse bool, err error)
}

// Keys service struct
type Service struct {
	devMode bool
	logger  *zap.SugaredLogger
	output  *output.Service
	storage Storage
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// devMode
	service.devMode = app.GetDevMode()

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
	service.storage = app.GetKeyStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	return service, nil
}
