package http01internal

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary http-01 internal challenge service component is missing")
	errConfigComponent  = errors.New("necessary http-01 config option missing")
)

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	devMode bool
	logger  *zap.SugaredLogger
	tokens  map[string]string
	mu      sync.RWMutex // added mutex due to unsafe if add and remove token both run
}

// Configuration options
type Config struct {
	Port *int `yaml:"port"`
}

// NewService creates a new service
func NewService(app App, config *Config) (*Service, error) {
	service := new(Service)

	// devmode?
	service.devMode = app.GetDevMode()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// allocate token map
	service.tokens = make(map[string]string, 50)

	// start web server for http01 challenges
	if config.Port == nil {
		return nil, errConfigComponent
	}
	err := service.startServer(*config.Port)
	if err != nil {
		return nil, err
	}

	return service, nil
}
