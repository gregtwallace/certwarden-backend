package challenges

import (
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
)

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

// addDomains adds all of the available domains from a provider to the domains
// configured in challenges
func (service *Service) addDomains(provider providerService) error {
	// add each domain name to providers map
	domainNames := provider.AvailableDomains()
	for _, domain := range domainNames {
		exists, _ := service.domainProviders.Add(domain, provider)
		if exists {
			return errMultipleSameDomain(domain)
		}
	}

	return nil
}
