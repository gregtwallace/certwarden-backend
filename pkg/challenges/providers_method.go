package challenges

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/validation"
	"strings"
)

// domainsLen returns the number of domains in providers (including wildcard if
// there is one)
func (ps *providers) domainsLen() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return len(ps.dP)
}

// hasDnsProvider returns true if any of the providers uses the dns-01
// challenge type
func (ps *providers) hasDnsProvider() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for provider := range ps.pD {
		if provider.AcmeChallengeType() == acme.ChallengeTypeDns01 {
			return true
		}
	}

	return false
}

// configs returns all of the providers' configurations
func (ps *providers) configs() ProvidersConfigs {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.cfgs
}

// addProvider adds the provider and all of its domains. if a domain already
// exists, an error is returned
func (ps *providers) addProvider(p providerService) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// providers domain names
	domainNames := p.AvailableDomains()

	// validate each domain name and add to providers map
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

// provider returns the providerService for the given acme Identifier. If
// there is no provider for the Identifier, an error is returned instead.
func (ps *providers) provider(identifier acme.Identifier) (providerService, error) {
	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return nil, errors.New("acme identifier is not dns type (challenges pkg can only solve dns type)")
	}

	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// check if identifier value ends in the domain
	for domain := range ps.dP {
		if strings.HasSuffix(identifier.Value, domain) {
			return ps.dP[domain], nil
		}
	}

	// if domain was not found, return wild provider if it exists
	p, exists := ps.dP["*"]
	if exists {
		return p, nil
	}

	return nil, fmt.Errorf("could not find a challenge provider for the specified identifier (%s; %s)", identifier.Type, identifier.Value)

}
