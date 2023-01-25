package dns_checker

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary challenges service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetShutdownContext() context.Context
	GetLogger() *zap.SugaredLogger
}

// Config is used to configure the service
type Config struct {
	SkipCheckWaitSeconds *int               `yaml:"skip_check_wait_seconds"`
	DnsServices          []DnsServiceIPPair `yaml:"dns_services"`
}

// service struct
type Service struct {
	shutdownContext context.Context
	logger          *zap.SugaredLogger
	skipWait        time.Duration
	dnsResolvers    []dnsResolverPair
}

// NewService creates a new service
func NewService(app App, cfg Config) (service *Service, err error) {
	service = new(Service)

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// configure resolvers (unless skipping check)
	if cfg.SkipCheckWaitSeconds != nil {
		service.logger.Warnf("dns record validation disabled, will manually sleep %d seconds instead", *cfg.SkipCheckWaitSeconds)
		service.skipWait = time.Duration(*cfg.SkipCheckWaitSeconds) * time.Second
	} else {
		service.dnsResolvers, err = makeResolvers(cfg.DnsServices)
		if err != nil {
			service.logger.Errorf("failed to configure dns checker resolvers (%s)", err)
			return nil, err
		}
	}

	return service, nil
}
