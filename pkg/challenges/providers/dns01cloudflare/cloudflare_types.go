package dns01cloudflare

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

// cloudflareResource returns the resource container for the specified resourceName.
// If a matching resource isn't found, an error is returned.
func (service *Service) cloudflareResource(resourceName string) (*cloudflare.ResourceContainer, error) {
	// look for domain suffix
	for domain := range service.domainIDs {
		if strings.HasSuffix(resourceName, domain) {
			return cloudflare.ZoneIdentifier(service.domainIDs[domain]), nil

		}
	}

	return nil, fmt.Errorf("dns01cloudflare could not find domain supporting resource name %s", resourceName)
}

// cloudflareCreateDNSParams returns the cloudflare create dns record params for a given
// acme resource name and content
func cloudflareCreateDNSParams(resourceName, resourceContent string) cloudflare.CreateDNSRecordParams {
	return cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    resourceName,
		Content: resourceContent,

		// specific to create
		TTL:       60,
		Proxiable: false,
		Comment:   fmt.Sprintf("created by LeGo CertHub on %s", time.Now().Format("Mon Jan 2 15:04:05 MST 2006")),
	}
}

// cloudflareListDNSParams returns the cloudflare list dns records params for a given
// acme resource name and content
func cloudflareListDNSParams(resourceName, resourceContent string) cloudflare.ListDNSRecordsParams {
	return cloudflare.ListDNSRecordsParams{
		Type:    "TXT",
		Name:    resourceName,
		Content: resourceContent,
	}
}
