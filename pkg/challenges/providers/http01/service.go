package http01

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary http-01 challenge service component is missing")

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

// NewService creates a new acme_accounts service
func NewService(app App, port int) (*Service, error) {
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
	err := service.startServer(port)
	if err != nil {
		return nil, err
	}

	return service, nil
}
