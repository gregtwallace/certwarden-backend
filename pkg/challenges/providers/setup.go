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
	"reflect"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// application contains functions needed in parent app
type application interface {
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetHttpClient() *httpclient.Client
	GetDevMode() bool
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Service is an interface for any provider service
type Service interface {
	AcmeChallengeType() acme.ChallengeType
	AvailableDomains() []string
	Provision(resourceName, resourceContent string) (err error)
	Deprovision(resourceName, resourceContent string) (err error)
	// Stop and Start are used in reconfiguring providers while app continues to run
	Stop() error
	Start() error
}

// provider is the structure of a provider that is being managed
type provider struct {
	ID      int    `json:"id"`
	Tag     string `json:"tag"`
	TypeOf  string `json:"type"`
	Config  any    `json:"config"`
	Service `json:"-"`
}

// Manager contains all providers and maintains several types
// of mappings between domains, providers, and provider types
type Manager struct {
	usable bool
	dP     map[string]*provider   // domain -> provider
	pD     map[*provider][]string // provider -> []domain
	tP     map[string][]*provider // typeOf -> []provider
	mu     sync.RWMutex
}

// addProvider adds the provider and all of its domains. if a domain already
// exists, an error is returned
func (mgr *Manager) addProvider(pService Service, cfg any) error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// no need to change usable; error or not of this func will not impact usability

	// parse config type to get provider type
	typeOf, _ := strings.CutPrefix(reflect.TypeOf(cfg).String(), "*")
	typeOf, cfgIsConfig := strings.CutSuffix(typeOf, ".Config")

	// create Provider from service and config
	p := &provider{
		ID:      len(mgr.pD),
		Tag:     randomness.GenerateInsecureString(10),
		TypeOf:  typeOf,
		Config:  cfg,
		Service: pService,
	}

	// always add provider to providers map, this ensures that if there is an error,
	// Stop() can still be called on all providers that were created
	mgr.pD[p] = []string{}
	mgr.tP[typeOf] = append(mgr.tP[typeOf], p)

	// if type of cfg was not a .Config, error
	if !cfgIsConfig {
		return errors.New("error adding provider, cfg is not a .Config (should never happen: report this as a lego application bug)")
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
		_, exists := mgr.dP[domain]
		if exists {
			return fmt.Errorf("failed to configure domain %s, each domain can only be configured once", domain)
		}

		// add to both internal maps
		mgr.dP[domain] = p
		mgr.pD[p] = append(mgr.pD[p], domain)
	}

	return nil
}

// MakeManager creates a manager of providers and creates all of the provider services
// using the specified cfg
func MakeManager(app application, cfg Config) (mgr *Manager, usesDns bool, err error) {
	// make struct with configs
	mgr = &Manager{
		usable: true,
		dP:     make(map[string]*provider),
		pD:     make(map[*provider][]string),
		tP:     make(map[string][]*provider),
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
			wgErrors <- mgr.addProvider(http01Internal, cfg.Http01InternalConfigs[i])
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
			wgErrors <- mgr.addProvider(dns01Manual, cfg.Dns01ManualConfigs[i])
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
			wgErrors <- mgr.addProvider(dns01AcmeDns, cfg.Dns01AcmeDnsConfigs[i])
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
			wgErrors <- mgr.addProvider(dns01AcmeSh, cfg.Dns01AcmeShConfigs[i])
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
			wgErrors <- mgr.addProvider(cloudflareProvider, cfg.Dns01CloudflareConfigs[i])
		}(i)
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			stopErr := mgr.Stop()
			if stopErr != nil {
				app.GetLogger().Fatalf("failed to stop challenge provider(s) (%s), fatal crash due to instability", stopErr)
				// app exits
			}
			return nil, false, err
		}
	}

	// verify at least one domain / provider exists
	if len(mgr.dP) <= 0 {
		return nil, false, errors.New("no challenge providers are properly configured (at least one must be enabled)")
	}

	// check for dns
	usesDns = false
	for provider := range mgr.pD {
		if provider.AcmeChallengeType() == acme.ChallengeTypeDns01 {
			usesDns = true
		}
	}

	return mgr, usesDns, nil
}
