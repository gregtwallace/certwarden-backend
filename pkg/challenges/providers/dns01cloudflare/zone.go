package dns01cloudflare

import (
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// zone stores information about a domain and is used
// to identify which API should be used and to prevent
// repeated unneeded ZoneID lookups via the API
type zone struct {
	id  string
	api *cloudflare.API
}

// getResourceZone returns the zone record for the specified
// resourceName
func (service *Service) getResourceZone(resourceName string) (zone, error) {
	// search for a known zone as the suffix of the resource name
	resourceZoneName := ""
	for zoneName := range service.knownDomainZones {
		// if resource ends in the zone name
		if strings.HasSuffix(resourceName, zoneName) {
			// found matching zone
			resourceZoneName = zoneName
			break
		}
	}

	// if no zone found
	if resourceZoneName == "" {
		return zone{}, errors.New("zone name does not exist in zone map")
	}

	// get the zone related to the resourceName
	z, exists := service.knownDomainZones[resourceZoneName]
	if !exists {
		// should be impossible after above name determination
		return zone{}, errors.New("zone name does not exist in zone map")
	}

	return z, nil
}
