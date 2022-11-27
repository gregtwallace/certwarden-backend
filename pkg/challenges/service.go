package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/http01internal"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary challenges service component is missing")
	errNoProviders      = errors.New("no challenge providers are properly configured (at least one must be enabled)")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetDevMode() bool
	GetHttp01InternalConfig() *http01internal.Config
	GetDns01CloudflareConfig() *dns01cloudflare.Config
}

// interface for any provider service
type providerService interface {
	Provision(resourceName string, resourceContent string) (err error)
	Deprovision(resourceName string, resourceContent string) (err error)
}

// service struct
type Service struct {
	logger      *zap.SugaredLogger
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	dnsChecker  *dns_checker.Service
	providers   map[MethodValue]providerService
	methods     []Method
}

// NewService creates a new service
func NewService(app App) (service *Service, err error) {
	service = new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// configure dns checker service
	// TODO: omit this if no dns providers are enabled
	service.dnsChecker, err = dns_checker.NewService(app)
	if err != nil {
		return nil, err
	}

	// challenge providers
	service.providers = make(map[MethodValue]providerService)

	// http-01 internal challenge server
	service.providers[methodValueHttp01Internal], err = http01internal.NewService(app, app.GetHttp01InternalConfig())
	if err != nil {
		return nil, err
	}

	// dns-01 cloudflare challenge service
	service.providers[methodValueDns01Cloudflare], err = dns01cloudflare.NewService(app, app.GetDns01CloudflareConfig(), service.dnsChecker)
	if err != nil {
		return nil, err
	}
	// end challenge providers

	// configure methods (list of all, properly flagged as enabled or not)
	service.configureMethods()

	// confirm at least one Method is enabled, else error
	atLeastOneMethod := false
	for i := range service.methods {
		if service.methods[i].Enabled {
			atLeastOneMethod = true
		}
	}
	if !atLeastOneMethod {
		return nil, errNoProviders
	}

	return service, nil
}
