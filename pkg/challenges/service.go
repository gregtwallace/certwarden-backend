package challenges

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/httpclient"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary challenges service component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetDevMode() bool
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// interface for any provider service
type providerService interface {
	Provision(resourceName string, resourceContent string) (err error)
	Deprovision(resourceName string, resourceContent string) (err error)
}

// ConfigProviders holds the challenge provider configs
type ConfigProviders struct {
	Http01InternalConfig  http01internal.Config  `yaml:"http_01_internal"`
	Dns01ManualConfig     dns01manual.Config     `yaml:"dns_01_manual"`
	Dns01CloudflareConfig dns01cloudflare.Config `yaml:"dns_01_cloudflare"`
	Dns01AcmeDnsConfig    dns01acmedns.Config    `yaml:"dns_01_acme_dns"`
}

// Config holds all of the challenge config
type Config struct {
	DnsCheckerConfig dns_checker.Config `yaml:"dns_checker"`
	ProviderConfigs  ConfigProviders    `yaml:"providers"`
}

// service struct
type Service struct {
	shutdownContext context.Context
	logger          *zap.SugaredLogger
	acmeProd        *acme.Service
	acmeStaging     *acme.Service
	dnsChecker      *dns_checker.Service
	providers       map[MethodValue]providerService
	methods         []Method
}

// NewService creates a new service
func NewService(app App, cfg *Config) (service *Service, err error) {
	service = new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// configure dns checker service (if any dns methods enabled)
	enableDnsChecker := cfg.ProviderConfigs.Dns01CloudflareConfig.Enable != nil && *cfg.ProviderConfigs.Dns01CloudflareConfig.Enable

	if enableDnsChecker {
		service.dnsChecker, err = dns_checker.NewService(app, cfg.DnsCheckerConfig)
		if err != nil {
			service.logger.Errorf("failed to configure dns checker (%s)", err)
			return nil, err
		}
	}

	// challenge providers
	service.providers = make(map[MethodValue]providerService)

	// http-01 internal challenge server
	http01Internal, err := http01internal.NewService(app, &cfg.ProviderConfigs.Http01InternalConfig)
	if err != nil {
		service.logger.Errorf("failed to configure http 01 internal (%s)", err)
		return nil, err
	}
	if http01Internal != nil {
		service.providers[methodValueHttp01Internal] = http01Internal
	}

	// dns-01 manual external scripts
	dns01Manual, err := dns01manual.NewService(app, &cfg.ProviderConfigs.Dns01ManualConfig, service.dnsChecker)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 manual (%s)", err)
		return nil, err
	}
	if dns01Manual != nil {
		service.providers[methodValueDns01Manual] = dns01Manual
	}

	// dns-01 cloudflare challenge service
	dns01Cloudflare, err := dns01cloudflare.NewService(app, &cfg.ProviderConfigs.Dns01CloudflareConfig, service.dnsChecker)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 cloudflare (%s)", err)
		return nil, err
	}
	if dns01Cloudflare != nil {
		service.providers[methodValueDns01Cloudflare] = dns01Cloudflare
	}

	// dns-01 acme-dns challenge service
	dns01AcmeDns, err := dns01acmedns.NewService(app, &cfg.ProviderConfigs.Dns01AcmeDnsConfig, service.dnsChecker)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 acme-dns (%s)", err)
		return nil, err
	}
	if dns01AcmeDns != nil {
		service.providers[methodValueDns01AcmeDns] = dns01AcmeDns
	}
	// end challenge providers

	// configure methods (list of all, properly flagged as enabled or not)
	err = service.configureMethods()
	if err != nil {
		service.logger.Errorf("failed to configure challenge methods (%s)", err)
		return nil, err
	}

	return service, nil
}
