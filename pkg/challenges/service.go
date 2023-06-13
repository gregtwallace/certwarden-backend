package challenges

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/httpclient"
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
	GetHttpClient() *httpclient.Client
	GetAcmeServerService() *acme_servers.Service
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
	Dns01AcmeDnsConfig    dns01acmedns.Config    `yaml:"dns_01_acme_dns"`
	Dns01AcmeShConfig     dns01acmesh.Config     `yaml:"dns_01_acme_sh"`
	Dns01CloudflareConfig dns01cloudflare.Config `yaml:"dns_01_cloudflare"`
}

// Config holds all of the challenge config
type Config struct {
	DnsCheckerConfig dns_checker.Config `yaml:"dns_checker"`
	ProviderConfigs  ConfigProviders    `yaml:"providers"`
}

// service struct
type Service struct {
	shutdownContext   context.Context
	logger            *zap.SugaredLogger
	acmeServerService *acme_servers.Service
	dnsChecker        *dns_checker.Service
	providers         map[MethodValue]providerService
	methods           []MethodWithStatus
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
	service.acmeServerService = app.GetAcmeServerService()
	if service.acmeServerService == nil {
		return nil, errServiceComponent
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
	dns01Manual, err := dns01manual.NewService(app, &cfg.ProviderConfigs.Dns01ManualConfig)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 manual (%s)", err)
		return nil, err
	}
	if dns01Manual != nil {
		service.providers[methodValueDns01Manual] = dns01Manual
	}

	// dns-01 acme-dns challenge service
	dns01AcmeDns, err := dns01acmedns.NewService(app, &cfg.ProviderConfigs.Dns01AcmeDnsConfig)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 acme-dns (%s)", err)
		return nil, err
	}
	if dns01AcmeDns != nil {
		service.providers[methodValueDns01AcmeDns] = dns01AcmeDns
	}

	// dns-01 acme.sh script service
	dns01AcmeSh, err := dns01acmesh.NewService(app, &cfg.ProviderConfigs.Dns01AcmeShConfig)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 acme.sh (%s)", err)
		return nil, err
	}
	if dns01AcmeSh != nil {
		service.providers[methodValueDns01AcmeSh] = dns01AcmeSh
	}

	// dns-01 cloudflare challenge service
	dns01Cloudflare, err := dns01cloudflare.NewService(app, &cfg.ProviderConfigs.Dns01CloudflareConfig)
	if err != nil {
		service.logger.Errorf("failed to configure dns 01 cloudflare (%s)", err)
		return nil, err
	}
	if dns01Cloudflare != nil {
		service.providers[methodValueDns01Cloudflare] = dns01Cloudflare
	}

	// end challenge providers

	// make array containing service methods and if they're enabled or disabled
	// fail out if none enabled
	atLeastOneEnabled := false
	for i := range allMethods {
		if _, ok := service.providers[allMethods[i].Value]; ok {
			// enabled
			service.methods = append(service.methods, allMethods[i].addStatus(true))
			atLeastOneEnabled = true
		} else {
			// disabled
			service.methods = append(service.methods, allMethods[i].addStatus(false))
		}
	}
	if !atLeastOneEnabled {
		return nil, errNoProviders
	}

	// configure dns checker service (if any enabled Method is a DNS method)
	// Fixes https://github.com/gregtwallace/legocerthub/issues/6
	for i := range service.methods {
		if service.methods[i].Enabled && service.methods[i].ChallengeType == acme.ChallengeTypeDns01 {
			// enable checker
			service.dnsChecker, err = dns_checker.NewService(app, cfg.DnsCheckerConfig)
			if err != nil {
				service.logger.Errorf("failed to configure dns checker (%s)", err)
				return nil, err
			}

			// no need to continue loop once DNS challenge service is confirmed required
			break
		}
	}

	return service, nil
}
