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
	http01Internal, err := http01internal.NewService(app, app.GetHttp01InternalConfig())
	if err != nil {
		return nil, err
	}
	if http01Internal != nil {
		service.providers[methodValueHttp01Internal] = http01Internal
	}

	// dns-01 cloudflare challenge service
	dns01Cloudflare, err := dns01cloudflare.NewService(app, app.GetDns01CloudflareConfig(), service.dnsChecker)
	if err != nil {
		return nil, err
	}
	if dns01Cloudflare != nil {
		service.providers[methodValueDns01Cloudflare] = dns01Cloudflare
	}
	// end challenge providers

	// configure methods (list of all, properly flagged as enabled or not)
	err = service.configureMethods()
	if err != nil {
		return nil, err
	}

	return service, nil
}
