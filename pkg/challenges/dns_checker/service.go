package dns_checker

import (
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary challenges service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Config is used to configure the service
type Config struct {
	DnsServices []DnsServiceIPPair `yaml:"dns_services"`
}

// service struct
type Service struct {
	logger       *zap.SugaredLogger
	dnsResolvers []dnsResolverPair
}

// NewService creates a new service
func NewService(app App, cfg Config) (service *Service, err error) {
	service = new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// configure resolvers
	service.dnsResolvers, err = makeResolvers(cfg.DnsServices)
	if err != nil {
		service.logger.Errorf("failed to configure dns checker resolvers (%s)", err)
		return nil, err
	}

	return service, nil
}
