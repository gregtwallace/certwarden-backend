package challenges

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/validation"
)

// ProviderConfigs provides structure for all provider config types
type ProvidersConfigs struct {
	Http01InternalConfigs  []http01internal.Config  `yaml:"http_01_internal"`
	Dns01ManualConfigs     []dns01manual.Config     `yaml:"dns_01_manual"`
	Dns01AcmeDnsConfigs    []dns01acmedns.Config    `yaml:"dns_01_acme_dns"`
	Dns01AcmeShConfigs     []dns01acmesh.Config     `yaml:"dns_01_acme_sh"`
	Dns01CloudflareConfigs []dns01cloudflare.Config `yaml:"dns_01_cloudflare"`
}

// Len returns the total number of Provider Configs, regardless of type.
func (cfg ProvidersConfigs) Len() int {
	return len(cfg.Dns01AcmeDnsConfigs) +
		len(cfg.Dns01AcmeShConfigs) +
		len(cfg.Dns01CloudflareConfigs) +
		len(cfg.Dns01ManualConfigs) +
		len(cfg.Http01InternalConfigs)
}

// Config holds all of the challenge config
type Config struct {
	DnsCheckerConfig dns_checker.Config `yaml:"dns_checker"`
	ProviderConfigs  ProvidersConfigs   `yaml:"providers"`
}

// addDomains adds all of the available domains from a provider to the
// challenges service
func (service *Service) addDomains(provider providerService) error {
	// providers domain names
	domainNames := provider.AvailableDomains()

	// check if provider is wild provider, if so, add wild card
	if len(domainNames) == 1 && domainNames[0] == "*" {
		err := service.domainProviders.add(domainNames[0], provider)
		if err != nil {
			return err
		}

		// done
		return nil
	}

	// if not wild card, validate each domain name and add to providers map
	for _, domain := range domainNames {

		// validate domain (wild with domain is never okay in challenges domain list)
		valid := validation.DomainValid(domain, false)
		if !valid {
			if domain == "*" {
				return errors.New("when using wildcard domain * it must be the only specified domain on the provider")
			}
			return fmt.Errorf("domain %s is not a validly formatted domain", domain)
		}

		err := service.domainProviders.add(domain, provider)
		if err != nil {
			return err
		}
	}

	return nil
}
