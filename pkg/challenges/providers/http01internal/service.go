package http01internal

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/datatypes"
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
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// provider Service struct
type Service struct {
	devMode              bool
	logger               *zap.SugaredLogger
	shutdownContext      context.Context
	shutdownWaitgroup    *sync.WaitGroup
	stopServerFunc       context.CancelFunc
	stopErrChan          chan error
	port                 int
	domains              []string
	provisionedResources *datatypes.SafeMap[[]byte]
}

// Stop stops the http server
func (service *Service) Stop() error {
	// shutdown http server
	service.stopServerFunc()

	// wait for result of server shutdown
	err := <-service.stopErrChan
	if err != nil {
		return err
	}

	return nil
}

// Start starts the http server
func (service *Service) Start() error {
	err := service.startServer()
	if err != nil {
		return err
	}

	return nil
}

// Configuration options
type Config struct {
	Domains []string `yaml:"domains" json:"domains"`
	Port    *int     `yaml:"port" json:"port"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config or no domains, error
	if cfg == nil || len(cfg.Domains) <= 0 {
		return nil, errServiceComponent
	}

	service := new(Service)

	// devmode?
	service.devMode = app.GetDevMode()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// set supported domains from config
	service.domains = append(service.domains, cfg.Domains...)

	// allocate resources map
	service.provisionedResources = datatypes.NewSafeMap[[]byte]()

	// set port
	if cfg.Port == nil {
		return nil, errConfigComponent
	}
	service.port = *cfg.Port

	// parent shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// parent shutdown wg
	service.shutdownWaitgroup = app.GetShutdownWaitGroup()

	// start web server for http01 challenges
	err := service.startServer()
	if err != nil {
		return nil, err
	}

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is http-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeHttp01
}
