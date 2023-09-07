package challenges

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/datatypes"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary challenges service component is missing")
	errNoProviders      = errors.New("no challenge providers are properly configured (at least one must be enabled)")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetOutputter() *output.Service

	// for provider children
	GetHttpClient() *httpclient.Client
	GetDevMode() bool
	GetShutdownWaitGroup() *sync.WaitGroup
}

// interface for any provider service
type providerService interface {
	AcmeChallengeType() acme.ChallengeType
	AvailableDomains() []string
	Provision(resourceName, resourceContent string) (err error)
	Deprovision(resourceName, resourceContent string) (err error)
}

// service struct
type Service struct {
	logger          *zap.SugaredLogger
	shutdownContext context.Context
	output          *output.Service
	dnsChecker      *dns_checker.Service
	providers       *providers
	resources       *datatypes.WorkTracker // tracks all resource names currently in use (regardless of provider)
}

// NewService creates a new service
func NewService(app App, cfg *Config) (service *Service, err error) {
	service = new(Service)

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
	service.providers, err = makeProviders(app, cfg.ProviderConfigs)
	if err != nil {
		service.logger.Errorf("failed to configure challenge provider(s) (%s)", err)
		return nil, err
	}
	if service.providers == nil {
		return nil, errServiceComponent
	}

	// verify at least one domain / provider exists
	if service.providers.domainsLen() <= 0 {
		return nil, errNoProviders
	}

	// configure dns checker service if any provider uses dns-01
	if service.providers.hasDnsProvider() {
		// enable checker
		service.dnsChecker, err = dns_checker.NewService(app, cfg.DnsCheckerConfig)
		if err != nil {
			service.logger.Errorf("failed to configure dns checker (%s)", err)
			return nil, err
		}
	}

	// make tracking map
	service.resources = datatypes.NewWorkTracker()

	return service, nil
}
