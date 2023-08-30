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

// Accounts service struct
type Service struct {
	devMode              bool
	logger               *zap.SugaredLogger
	domains              []string
	provisionedResources *datatypes.SafeMap[[]byte]
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

	// start web server for http01 challenges
	if cfg.Port == nil {
		return nil, errConfigComponent
	}
	err := service.startServer(*cfg.Port, app.GetShutdownContext(), app.GetShutdownWaitGroup())
	if err != nil {
		return nil, err
	}

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is http-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeHttp01
}
