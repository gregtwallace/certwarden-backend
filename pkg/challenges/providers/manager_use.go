package providers

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"strings"
)

// ProviderFor returns the provider Service for the given acme Identifier. If
// there is no provider for the Identifier, an error is returned instead.
func (mgr *Manager) ProviderFor(identifier acme.Identifier) (*provider, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return nil, errors.New("acme identifier is not dns type (challenges pkg can only solve dns type)")
	}

	// if exact domain is in the list, return its provider
	p, exists := mgr.dP[identifier.Value]
	if exists {
		return p, nil
	}

	// find best match from options (if there is a provider for a more specific subdomain, choose that one)
	providerDomain := ""
	for domain := range mgr.dP {
		// include period to avoid matching something like hellodomain.com to domain.com 's provider
		if strings.HasSuffix(identifier.Value, "."+domain) {
			// for a provider with the proper suffix, check length of existing match and update
			// match if the new match is longer
			if len(domain) > len(providerDomain) {
				providerDomain = domain
			}
		}
	}
	// if a match was found, return its provider
	if providerDomain != "" {
		return mgr.dP[providerDomain], nil
	}

	// if domain was not found, return wild provider if it exists
	p, exists = mgr.dP["*"]
	if exists {
		return p, nil
	}

	return nil, fmt.Errorf("could not find a challenge provider for the specified identifier (%s; %s)", identifier.Type, identifier.Value)

}
