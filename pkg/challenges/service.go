package challenges

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers"
	"legocerthub-backend/pkg/datatypes"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary challenges service component is missing")
)

// App interface is for connecting to the main app
type application interface {
	GetConfigFilename() string
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetOutputter() *output.Service

	// for providers
	GetHttpClient() *httpclient.Client
	GetDevMode() bool
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Config holds all of the challenge config
type Config struct {
	DnsCheckerConfig dns_checker.Config `yaml:"dns_checker"`
	ProviderConfigs  providers.Config   `yaml:"providers"`
}

// service struct
type Service struct {
	app             application
	dnsCheckerCfg   dns_checker.Config
	logger          *zap.SugaredLogger
	shutdownContext context.Context
	output          *output.Service
	dnsChecker      *dns_checker.Service
	Providers       *providers.Manager
	resources       *datatypes.WorkTracker // tracks all resource names currently in use (regardless of provider)
}

// NewService creates a new service
func NewService(app application, cfg *Config) (service *Service, err error) {
	service = new(Service)

	// save app pointer for use later in reconfiguring providers
	service.app = app

	// save dns checker config for use later in reconfiguring providers
	service.dnsCheckerCfg = cfg.DnsCheckerConfig

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output
	service.output = app.GetOutputter()

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// configure challenge providers
	service.Providers, err = providers.MakeManager(app, cfg.ProviderConfigs)
	if err != nil {
		service.logger.Errorf("failed to configure challenge provider(s) (%s)", err)
		return nil, err
	}

	// create dns checker regardless of if using dns (since providers can change)
	service.dnsChecker, err = dns_checker.NewService(app, cfg.DnsCheckerConfig)
	if err != nil {
		service.logger.Errorf("failed to configure dns checker (%s)", err)
		return nil, err
	}

	// make tracking map
	service.resources = datatypes.NewWorkTracker()

	return service, nil
}
