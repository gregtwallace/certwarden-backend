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
	"legocerthub-backend/pkg/randomness"
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
// mapping between all domains and all providers and a mapping from id to provider
type Providers struct {
	usable bool
	cfg    Config
	dP     map[string]Service   // domain -> provider
	pD     map[Service][]string // provider -> []domain
	idP    map[int]Service      // id -> provider
	mu     sync.RWMutex
}

// addProvider adds the provider and all of its domains. if a domain already
// exists, an error is returned
func (ps *Providers) addProvider(p Service, cfg providerService) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// no need to change usable; error or not of this func will not impact usability

	// always add provider to providers map, this ensures that if there is an error,
	// Stop() can still be called on all providers that were created

	// store provider and its id
	id := len(ps.pD)
	ps.idP[id] = p
	ps.pD[p] = []string{}

	// store provider config with ID and Tag
	cfg.SetIDAndTag(id, randomness.GenerateInsecureString(10))
	switch cfg := cfg.(type) {
	case *http01internal.Config:
		ps.cfg.Http01InternalConfigs = append(ps.cfg.Http01InternalConfigs, cfg)
	case *dns01manual.Config:
		ps.cfg.Dns01ManualConfigs = append(ps.cfg.Dns01ManualConfigs, cfg)
	case *dns01acmedns.Config:
		ps.cfg.Dns01AcmeDnsConfigs = append(ps.cfg.Dns01AcmeDnsConfigs, cfg)
	case *dns01acmesh.Config:
		ps.cfg.Dns01AcmeShConfigs = append(ps.cfg.Dns01AcmeShConfigs, cfg)
	case *dns01cloudflare.Config:
		ps.cfg.Dns01CloudflareConfigs = append(ps.cfg.Dns01CloudflareConfigs, cfg)
	default:
		return errors.New("failed adding provider, config unsupported (should never happen, report as bug)")
	}

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
		dP:     make(map[string]Service),
		pD:     make(map[Service][]string),
		idP:    make(map[int]Service),
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
			http01Internal, err := http01internal.NewService(app, cfg.Http01InternalConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(http01Internal, cfg.Http01InternalConfigs[i])
		}(i)
	}

	// dns-01 manual external script services
	for i := range cfg.Dns01ManualConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01Manual, err := dns01manual.NewService(app, cfg.Dns01ManualConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01Manual, cfg.Dns01ManualConfigs[i])
		}(i)
	}

	// dns-01 acme-dns challenge services
	for i := range cfg.Dns01AcmeDnsConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeDns, err := dns01acmedns.NewService(app, cfg.Dns01AcmeDnsConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeDns, cfg.Dns01AcmeDnsConfigs[i])
		}(i)
	}

	// dns-01 acme.sh script services
	for i := range cfg.Dns01AcmeShConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeSh, err := dns01acmesh.NewService(app, cfg.Dns01AcmeShConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeSh, cfg.Dns01AcmeShConfigs[i])
		}(i)
	}

	// dns-01 cloudflare challenge services
	for i := range cfg.Dns01CloudflareConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			cloudflareProvider, err := dns01cloudflare.NewService(app, cfg.Dns01CloudflareConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(cloudflareProvider, cfg.Dns01CloudflareConfigs[i])
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
