package dns01cloudflare

import (
	"errors"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/datatypes"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	logger           *zap.SugaredLogger
	dnsChecker       *dns_checker.Service
	knownDomainZones *datatypes.SafeMap
	dnsRecords       *datatypes.SafeMap
}

// NewService creates a new service
func NewService(app App, config *Config, dnsChecker *dns_checker.Service) (*Service, error) {
	// if disabled, return nil and no error
	if !*config.Enable {
		return nil, nil
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// dns checker
	service.dnsChecker = dnsChecker

	// cloudflare api
	err := service.configureCloudflareAPI(config)
	if err != nil {
		return nil, err
	}

	// debug log configured domains
	service.logger.Debugf("dns01cloudflare configured domains: %s", service.knownDomainZones.ListKeys())

	// map to hold current dnsRecords
	service.dnsRecords = datatypes.NewSafeMap()

	return service, nil
}
