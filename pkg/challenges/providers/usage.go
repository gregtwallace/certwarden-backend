package providers

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"strings"
)

// Provider returns the providerService for the given acme Identifier. If
// there is no provider for the Identifier, an error is returned instead.
func (ps *Providers) Provider(identifier acme.Identifier) (Service, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// check that providers is usable
	if !ps.usable {
		return nil, errors.New("providers not currently in usable state")
	}

	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return nil, errors.New("acme identifier is not dns type (challenges pkg can only solve dns type)")
	}

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
