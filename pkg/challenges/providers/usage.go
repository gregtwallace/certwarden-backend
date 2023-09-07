package providers

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"strings"
)

var errMgrUnusable = errors.New("providers manager not currently in usable state")

// Provider returns the provider with the specified id. if no such provider
// exists, an error is returned instead
func (mgr *Manager) Provider(id int) (*provider, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// fail if not usable
	if !mgr.usable {
		return nil, errMgrUnusable
	}

	p, exists := mgr.iP[id]
	if !exists {
		return nil, fmt.Errorf("no provider exists with id %d", id)
	}

	return p, nil
}

// Providers returns all of the providers in manager, grouped by typeOf
func (mgr *Manager) Providers() (map[string][]*provider, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// fail if not usable
	if !mgr.usable {
		return nil, errMgrUnusable
	}

	return mgr.tP, nil
}

// ProviderFor returns the provider Service for the given acme Identifier. If
// there is no provider for the Identifier, an error is returned instead.
func (mgr *Manager) ProviderFor(identifier acme.Identifier) (Service, error) {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// fail if not usable
	if !mgr.usable {
		return nil, errMgrUnusable
	}

	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return nil, errors.New("acme identifier is not dns type (challenges pkg can only solve dns type)")
	}

	// check if identifier value ends in the domain
	for domain := range mgr.dP {
		if strings.HasSuffix(identifier.Value, domain) {
			return mgr.dP[domain], nil
		}
	}

	// if domain was not found, return wild provider if it exists
	p, exists := mgr.dP["*"]
	if exists {
		return p, nil
	}

	return nil, fmt.Errorf("could not find a challenge provider for the specified identifier (%s; %s)", identifier.Type, identifier.Value)

}
