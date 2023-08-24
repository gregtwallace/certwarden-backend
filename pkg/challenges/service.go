package challenges

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/datatypes"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/httpclient"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent   = errors.New("necessary challenges service component is missing")
	errNoProviders        = errors.New("no challenge providers are properly configured (at least one must be enabled)")
	errMultipleSameDomain = func(domainName string) error {
		return fmt.Errorf("failed to configure domain %s, each domain can only be configured once", domainName)
	}
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
	AcmeChallengeType() acme.ChallengeType
	AvailableDomains() []string
	Provision(domainName, resourceName, resourceContent string) (err error)
	Deprovision(domainName, resourceName, resourceContent string) (err error)
}

// ProviderConfigs holds the challenge provider configs
type ProviderConfigs struct {
	Http01InternalConfigs  []http01internal.Config  `yaml:"http_01_internal"`
	Dns01ManualConfigs     []dns01manual.Config     `yaml:"dns_01_manual"`
	Dns01AcmeDnsConfigs    []dns01acmedns.Config    `yaml:"dns_01_acme_dns"`
	Dns01AcmeShConfigs     []dns01acmesh.Config     `yaml:"dns_01_acme_sh"`
	Dns01CloudflareConfigs []dns01cloudflare.Config `yaml:"dns_01_cloudflare"`
}

// Config holds all of the challenge config
type Config struct {
	DnsCheckerConfig dns_checker.Config `yaml:"dns_checker"`
	ProviderConfigs  ProviderConfigs    `yaml:"providers"`
}

// service struct
type Service struct {
	shutdownContext    context.Context
	logger             *zap.SugaredLogger
	acmeServerService  *acme_servers.Service
	dnsChecker         *dns_checker.Service
	domainProviders    *datatypes.SafeMap[providerService] // domain_name[providerService]
	resourceNamesInUse *datatypes.WorkTracker              // tracks all resource names currently in use (regardless of provider)
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
	service.domainProviders = datatypes.NewSafeMap[providerService]()

	// // http-01 internal challenge server
	// http01Internal, err := http01internal.NewService(app, &cfg.ProviderConfigs.Http01InternalConfig)
	// if err != nil {
	// 	service.logger.Errorf("failed to configure http 01 internal (%s)", err)
	// 	return nil, err
	// }
	// if http01Internal != nil {
	// 	service.providers[methodValueHttp01Internal] = http01Internal
	// }

	// // dns-01 manual external scripts
	// dns01Manual, err := dns01manual.NewService(app, &cfg.ProviderConfigs.Dns01ManualConfig)
	// if err != nil {
	// 	service.logger.Errorf("failed to configure dns 01 manual (%s)", err)
	// 	return nil, err
	// }
	// if dns01Manual != nil {
	// 	service.providers[methodValueDns01Manual] = dns01Manual
	// }

	// // dns-01 acme-dns challenge service
	// dns01AcmeDns, err := dns01acmedns.NewService(app, &cfg.ProviderConfigs.Dns01AcmeDnsConfig)
	// if err != nil {
	// 	service.logger.Errorf("failed to configure dns 01 acme-dns (%s)", err)
	// 	return nil, err
	// }
	// if dns01AcmeDns != nil {
	// 	service.providers[methodValueDns01AcmeDns] = dns01AcmeDns
	// }

	// // dns-01 acme.sh script service
	// dns01AcmeSh, err := dns01acmesh.NewService(app, &cfg.ProviderConfigs.Dns01AcmeShConfig)
	// if err != nil {
	// 	service.logger.Errorf("failed to configure dns 01 acme.sh (%s)", err)
	// 	return nil, err
	// }
	// if dns01AcmeSh != nil {
	// 	service.providers[methodValueDns01AcmeSh] = dns01AcmeSh
	// }

	// dns-01 cloudflare challenge services
	for i := range cfg.ProviderConfigs.Dns01CloudflareConfigs {
		cloudflareProvider, err := dns01cloudflare.NewService(app, &cfg.ProviderConfigs.Dns01CloudflareConfigs[i])
		if err != nil || cloudflareProvider == nil {
			service.logger.Errorf("failed to configure cloudflare challenge provider instance (%s)", err)
			return nil, err
		}

		// add each domain name to providers map
		domainNames := cloudflareProvider.AvailableDomains()
		for _, domain := range domainNames {
			exists, _ := service.domainProviders.Add(domain, cloudflareProvider)
			if exists {
				return nil, errMultipleSameDomain(domain)
			}
		}
	}

	// end challenge providers

	// verify at least one domain / provider exists
	if service.domainProviders.Len() <= 0 {
		return nil, errNoProviders
	}

	// configure dns checker service (if any domain uses a dns-01 provider)
	checkDns01ProviderFunc := func(p providerService) bool {
		return p.AcmeChallengeType() == acme.ChallengeTypeDns01
	}

	if service.domainProviders.CheckValuesForFunc(checkDns01ProviderFunc) {
		// enable checker
		service.dnsChecker, err = dns_checker.NewService(app, cfg.DnsCheckerConfig)
		if err != nil {
			service.logger.Errorf("failed to configure dns checker (%s)", err)
			return nil, err
		}
	}

	// make tracking map
	service.resourceNamesInUse = datatypes.NewWorkTracker()

	return service, nil
}
