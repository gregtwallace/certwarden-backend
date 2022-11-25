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

// service struct
type Service struct {
	logger       *zap.SugaredLogger
	dnsResolvers []dnsResolverPair
}

// NewService creates a new service
func NewService(app App) (service *Service, err error) {
	service = new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// Set up dns IP pairs
	// TODO: Make dynamic in config
	dnsServices := []dnsServiceIPPair{
		// Cloudflare
		{
			primary:   "1.1.1.1",
			secondary: "1.0.0.1",
		},
		// Quad9
		{
			primary:   "9.9.9.9",
			secondary: "149.112.112.112",
		},
		// Google
		{
			primary:   "8.8.8.8",
			secondary: "4.4.4.4",
		},
	}

	// configure resolvers
	service.dnsResolvers = makeResolvers(dnsServices)

	return service, nil
}
