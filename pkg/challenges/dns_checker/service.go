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

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}
	service.logger.Debug("dns_checker: starting service")

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// configure resolvers (unless skipping check)
	if cfg.SkipCheckWaitSeconds != nil {
		service.logger.Warnf("dns checker: dns record validation disabled, will manually sleep %d seconds instead", *cfg.SkipCheckWaitSeconds)
		service.skipWait = time.Duration(*cfg.SkipCheckWaitSeconds) * time.Second
	} else {
		service.dnsResolvers, err = makeResolvers(cfg.DnsServices)
		if err != nil {
			// if failed to make resolvers, fallback to sleeping
			fallbackSleepSeconds := 120
			service.logger.Errorf("dns checker: failed to configure resolvers (%s), will sleep %d seconds instead of validating dns records", err, fallbackSleepSeconds)
			service.skipWait = time.Duration(fallbackSleepSeconds) * time.Second
		} else {
			// success
			service.logger.Debugf("dns checker: configured dns server pairs: %s", cfg.DnsServices)
		}
	}

	return service, nil
}
