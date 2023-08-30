package challenges

import (
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"sync"
)

// providers contains the configuration for all providers as well as a bi directional
// mapping between all domains and all providers
type providers struct {
	cfgs ProvidersConfigs
	dP   map[string]providerService
	pD   map[providerService][]string
	mu   sync.RWMutex
}

// makeProviders creates the providers struct and creates all of the provider services
// using the specified cfgs
func makeProviders(app App, configs ProvidersConfigs) (*providers, error) {
	// make struct with configs
	ps := &providers{
		cfgs: configs,
		dP:   make(map[string]providerService),
		pD:   make(map[providerService][]string),
	}

	// populate

	// configure providers async
	var wg sync.WaitGroup
	wgSize := ps.cfgs.Len()

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// http-01 internal challenge servers
	for i := range ps.cfgs.Http01InternalConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			http01Internal, err := http01internal.NewService(app, &ps.cfgs.Http01InternalConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(http01Internal)
		}(i)
	}

	// dns-01 manual external script services
	for i := range ps.cfgs.Dns01ManualConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01Manual, err := dns01manual.NewService(app, &ps.cfgs.Dns01ManualConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01Manual)
		}(i)
	}

	// dns-01 acme-dns challenge services
	for i := range ps.cfgs.Dns01AcmeDnsConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeDns, err := dns01acmedns.NewService(app, &ps.cfgs.Dns01AcmeDnsConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeDns)
		}(i)
	}

	// dns-01 acme.sh script services
	for i := range ps.cfgs.Dns01AcmeShConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			dns01AcmeSh, err := dns01acmesh.NewService(app, &ps.cfgs.Dns01AcmeShConfigs[i])
			if err != nil {
				wgErrors <- err
				return
			}

			// add each domain name to providers map
			wgErrors <- ps.addProvider(dns01AcmeSh)
		}(i)
	}

	// dns-01 cloudflare challenge services
	for i := range ps.cfgs.Dns01CloudflareConfigs {
		go func(i int) {
			// done after func
			defer wg.Done()

			// make service
			cloudflareProvider, err := dns01cloudflare.NewService(app, &ps.cfgs.Dns01CloudflareConfigs[i])
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
			return nil, err
		}
	}

	return ps, nil
}
