package challenges

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"strings"
	"sync"
)

// domainMap is a bi directional map that challenges uses since sometimes
// domains need to be looked up and other times a provider needs to be
// looked up
type domainProviderMap struct {
	dP map[string]providerService
	pD map[providerService][]string
	mu sync.RWMutex
}

// newDomainProviderMap creates a domainProviderMap to store domains and
// providerServices
func newDomainProviderMap() *domainProviderMap {
	return &domainProviderMap{
		dP: make(map[string]providerService),
		pD: make(map[providerService][]string),
	}
}

// CountDomains returns the number of domains contained in domainProviderMap
func (dpm *domainProviderMap) countDomains() int {
	dpm.mu.RLock()
	defer dpm.mu.RUnlock()

	return len(dpm.dP)
}

// hasDnsProvider returns true if any of the providers in domainProviderMap
// uses the dns-01 challenge type
func (dpm *domainProviderMap) hasDnsProvider() bool {
	dpm.mu.RLock()
	defer dpm.mu.RUnlock()

	for provider := range dpm.pD {
		if provider.AcmeChallengeType() == acme.ChallengeTypeDns01 {
			return true
		}
	}

	return false
}

// Add adds the domain to the domainProviderMap and sets its value to the
// providerService specified. If the domain already exists, an error is
// returned.
func (dpm *domainProviderMap) add(domain string, p providerService) error {
	dpm.mu.Lock()
	defer dpm.mu.Unlock()

	// if domain exists, return an error
	_, exists := dpm.dP[domain]
	if exists {
		return fmt.Errorf("failed to configure domain %s, each domain can only be configured once", domain)
	}

	// add the domain to both internal maps
	dpm.dP[domain] = p
	dpm.pD[p] = append(dpm.pD[p], domain)

	return nil
}

// getProvider returns the providerService for a given acme Identifier. If
// there is no provider for the Identifier, an error is returned instead.
func (dpm *domainProviderMap) getProvider(identifier acme.Identifier) (providerService, error) {
	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return nil, errors.New("acme identifier is not dns type (challenges pkg can only solve dns type)")
	}

	dpm.mu.RLock()
	defer dpm.mu.RUnlock()

	// check if identifier value ends in the domain
	for domain := range dpm.dP {
		if strings.HasSuffix(identifier.Value, domain) {
			return dpm.dP[domain], nil
		}
	}

	// if domain was not found, return wild provider if it exists
	p, exists := dpm.dP["*"]
	if exists {
		return p, nil
	}

	return nil, fmt.Errorf("could not find a challenge provider for the specified identifier (%s; %s)", identifier.Type, identifier.Value)

}
