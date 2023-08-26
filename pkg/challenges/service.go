package challenges

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
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

	// challenge providers - begin
	service.domainProviders = datatypes.NewSafeMap[providerService]()

	// configure providers async
	var wg sync.WaitGroup
	wgSize := cfg.ProviderConfigs.Len()

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// http-01 internal challenge servers
	for i := range cfg.ProviderConfigs.Http01InternalConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			http01Internal, err := http01internal.NewService(app, &cfg.ProviderConfigs.Http01InternalConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- service.addDomains(http01Internal)
		}(i)
	}

	// dns-01 manual external scripts
	for i := range cfg.ProviderConfigs.Dns01ManualConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01Manual, err := dns01manual.NewService(app, &cfg.ProviderConfigs.Dns01ManualConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- service.addDomains(dns01Manual)
		}(i)
	}

	// // dns-01 acme-dns challenge service
	// dns01AcmeDns, err := dns01acmedns.NewService(app, &cfg.ProviderConfigs.Dns01AcmeDnsConfig)
	// if err != nil {
	// 	service.logger.Errorf("failed to configure dns 01 acme-dns (%s)", err)
	// 	return nil, err
	// }
	// if dns01AcmeDns != nil {
	// 	service.providers[methodValueDns01AcmeDns] = dns01AcmeDns
	// }

	// dns-01 acme.sh script services
	for i := range cfg.ProviderConfigs.Dns01AcmeShConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeSh, err := dns01acmesh.NewService(app, &cfg.ProviderConfigs.Dns01AcmeShConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- service.addDomains(dns01AcmeSh)
		}(i)
	}

	// dns-01 cloudflare challenge services
	for i := range cfg.ProviderConfigs.Dns01CloudflareConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			cloudflareProvider, err := dns01cloudflare.NewService(app, &cfg.ProviderConfigs.Dns01CloudflareConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- service.addDomains(cloudflareProvider)
		}(i)
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			service.logger.Errorf("failed to configure challenge provider(s) (%s)", err)
			return nil, err
		}
	}

	// challenge providers - end

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
