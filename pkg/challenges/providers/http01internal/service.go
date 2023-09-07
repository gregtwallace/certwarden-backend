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
	provisionedResources *datatypes.SafeMap[[]byte]
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is http-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeHttp01
}

// Configuration options
type Config struct {
	Doms []string `yaml:"domains" json:"domains"`
	Port *int     `yaml:"port" json:"port"`
}

// Domains returns all of the domains specified in the Config
func (cfg *Config) Domains() []string {
	return cfg.Doms
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	if cfg == nil {
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

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) error {
	// old Service http server must be stopped before new service is created
	service.stopServerFunc()

	// wait for result of server shutdown
	err := <-service.stopErrChan
	if err != nil {
		return err
	}

	// make new service
	newServ, err := NewService(app, cfg)
	if err != nil {
		// if failed to make, restart old server
		errRestart := service.startServer()
		if errRestart != nil {
			service.logger.Fatalf("failed to restart http 01 server leaving http 01 internal provider in an unstable state")
			// ^ app terminates
			return errRestart
		}
		return err
	}

	// set content of old pointer so anything with the pointer calls the
	// updated service
	*service = *newServ

	return nil
}
