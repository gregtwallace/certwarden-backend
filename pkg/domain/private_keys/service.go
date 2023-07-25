package private_keys

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary key service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
	IsHttps() bool
	GetOutputter() *output.Service
	GetKeyStorage() Storage
}

// Storage interface for storage functions
type Storage interface {
	GetAllKeys(q pagination_sort.Query) (keys []Key, totalRows int, err error)
	GetOneKeyById(id int) (Key, error)
	GetOneKeyByName(name string) (Key, int, error)
	GetKeyPemById(id int) (name string, pem string, createdAt int, err error)

	PostNewKey(NewPayload) (keyId int, err error)

	PutKeyUpdate(UpdatePayload) error
	PutKeyApiKey(keyId int, apiKey string, updateTimeUnix int) (err error)
	PutKeyNewApiKey(keyId int, newApiKey string, updateTimeUnix int) error

	DeleteKey(int) error

	GetAvailableKeys() ([]Key, error)
	KeyInUse(id int) (inUse bool, err error)
}

// Keys service struct
type Service struct {
	devMode bool
	logger  *zap.SugaredLogger
	https   bool
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

	// running as https?
	service.https = app.IsHttps()

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
