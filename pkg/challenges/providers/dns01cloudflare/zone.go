package dns01cloudflare

import (
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// zone stores information about a domain and is used
// to identify which API should be used and to prevent
// repeated unneeded ZoneID lookups view the API
type zone struct {
	id  string
	api *cloudflare.API
}

// zoneName returns the domain name with TLD from the resourceName
func zoneName(resourceName string) (domain string) {
	domainParts := strings.Split(resourceName, ".")
	return domainParts[len(domainParts)-2] + "." + domainParts[len(domainParts)-1]
}

// getResourceZone returns the zone record for the specified
// resourceName
func (service *Service) getResourceZone(resourceName string) (zone, error) {
	// get the zone related to the resourceName
	z, exists := service.knownDomainZones[zoneName(resourceName)]
	if !exists {
		// MAYBE TODO: Check accounts again to see if zone was added. As built, LeGo will
		// need a restart if zones are added to an account
		return zone{}, errors.New("zone name does not exist in zone map")
	}

	return z, nil
}
