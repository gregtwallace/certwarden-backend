package providers

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/validation"
	"sync"

	"go.uber.org/zap"
)

type application interface {
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetHttpClient() *httpclient.Client
	GetDevMode() bool
	GetShutdownWaitGroup() *sync.WaitGroup
}

// interface for any provider service
type Service interface {
	AcmeChallengeType() acme.ChallengeType
	AvailableDomains() []string
	Provision(resourceName, resourceContent string) (err error)
	Deprovision(resourceName, resourceContent string) (err error)
	// Stop and Start are used in reconfiguring providers while app continues to run
	Stop() error
	Start() error
}

// providers contains the configuration for all providers as well as a bi directional
// mapping between all domains and all providers
type Providers struct {
	usable bool
	cfg    Config
	dP     map[string]Service
	pD     map[Service][]string
	mu     sync.RWMutex
}

// addProvider adds the provider and all of its domains. if a domain already
// exists, an error is returned
func (ps *Providers) addProvider(p Service) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// no need to change usable; error or not of this func will not impact usability

	// always add provider to providers map, this ensures that if there is an error,
	// Stop() can still be called on all providers that were created
	ps.pD[p] = []string{}

	// providers domain names
	domainNames := p.AvailableDomains()

	// add each domain to providers map
	for _, domain := range domainNames {
		// if not valid and not wild card service provider
		if !validation.DomainValid(domain, false) && !(len(domainNames) == 1 && domainNames[0] == "*") {
			if domain == "*" {
				return errors.New("when using wildcard domain * it must be the only specified domain on the provider")
			}
			return fmt.Errorf("domain %s is not a validly formatted domain", domain)
		}

		// if already exists, return an error
		_, exists := ps.dP[domain]
		if exists {
			return fmt.Errorf("failed to configure domain %s, each domain can only be configured once", domain)
		}

		// add to both internal maps
		ps.dP[domain] = p
		ps.pD[p] = append(ps.pD[p], domain)
	}

	return nil
}

// MakeProviders creates the providers struct and creates all of the provider services
// using the specified cfg
func MakeProviders(app application, cfg Config) (ps *Providers, usesDns bool, err error) {
	// make struct with configs
	ps = &Providers{
		usable: true,
		cfg:    cfg,
		dP:     make(map[string]Service),
		pD:     make(map[Service][]string),
	}

	// configure providers async
	var wg sync.WaitGroup
	wgSize := cfg.Len()

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// http-01 internal challenge servers
	for i := range cfg.Http01InternalConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			http01Internal, err := http01internal.NewService(app, &cfg.Http01InternalConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(http01Internal)
		}(i)
	}

	// dns-01 manual external script services
	for i := range cfg.Dns01ManualConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01Manual, err := dns01manual.NewService(app, &cfg.Dns01ManualConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01Manual)
		}(i)
	}

	// dns-01 acme-dns challenge services
	for i := range cfg.Dns01AcmeDnsConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeDns, err := dns01acmedns.NewService(app, &cfg.Dns01AcmeDnsConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeDns)
		}(i)
	}

	// dns-01 acme.sh script services
	for i := range cfg.Dns01AcmeShConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeSh, err := dns01acmesh.NewService(app, &cfg.Dns01AcmeShConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeSh)
		}(i)
	}

	// dns-01 cloudflare challenge services
	for i := range cfg.Dns01CloudflareConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			cloudflareProvider, err := dns01cloudflare.NewService(app, &cfg.Dns01CloudflareConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(cloudflareProvider)
		}(i)
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			stopErr := ps.Stop()
			if stopErr != nil {
				app.GetLogger().Fatalf("failed to stop challenge provider(s) (%s), fatal crash due to instability", stopErr)
				// app exits
			}
			return nil, false, err
		}
	}

	// verify at least one domain / provider exists
	if len(ps.dP) <= 0 {
		return nil, false, errors.New("no challenge providers are properly configured (at least one must be enabled)")
	}

	// check for dns
	usesDns = false
	for provider := range ps.pD {
		if provider.AcmeChallengeType() == acme.ChallengeTypeDns01 {
			usesDns = true
		}
	}

	return ps, usesDns, nil
}
