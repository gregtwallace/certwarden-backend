package dns01cloudflare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

// cloudflareResource returns the resource container for the specified dns record name.
// If a matching resource isn't found, an error is returned.
func (service *Service) cloudflareResource(dnsRecordName string) (*cloudflare.ResourceContainer, error) {
	// fetch list of zones
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	availableZones, err := service.cloudflareApi.ListZones(ctx)
	if err != nil {
		err = fmt.Errorf("dns01cloudflare api instance %s failed to list zones while searching for zone for %s (%s)", service.redactedApiIdentifier(), dnsRecordName, err)
		service.logger.Error(err)
		return nil, err
	}

	// find the zone for this resource
	resourceZone := cloudflare.Zone{}
	for i := range availableZones {
		// check if zone name is the suffix of resource name (i.e. this is the correct zone)
		if strings.HasSuffix(dnsRecordName, availableZones[i].Name) {
			resourceZone = availableZones[i]
			break
		}
	}
	// defer err check to after perm (zone not found won't have needed permission)

	// verify proper permission
	properPermission := false
	for i := range resourceZone.Permissions {
		if resourceZone.Permissions[i] == "#dns_records:edit" {
			properPermission = true
			break
		}
	}
	if !properPermission {
		return nil, fmt.Errorf("dns01cloudflare could not find cloudflare zone with proper permission supporting resource name %s", dnsRecordName)
	}

	return cloudflare.ZoneIdentifier(resourceZone.ID), nil
}

// cloudflareCreateDNSParams returns the cloudflare create dns record params for a given
// acme resource name and content
func cloudflareCreateDNSParams(dnsRecordName, dnsRecordValue string) cloudflare.CreateDNSRecordParams {
	return cloudflare.CreateDNSRecordParams{
		Type:    "TXT",
		Name:    dnsRecordName,
		Content: dnsRecordValue,

		// specific to create
		TTL:       60,
		Proxiable: false,
		Comment:   fmt.Sprintf("created by LeGo CertHub on %s", time.Now().Format("Mon Jan 2 15:04:05 MST 2006")),
	}
}

// cloudflareListDNSParams returns the cloudflare list dns records params for a given
// acme resource name and content
func cloudflareListDNSParams(dnsRecordName, dnsRecordValue string) cloudflare.ListDNSRecordsParams {
	return cloudflare.ListDNSRecordsParams{
		Type:    "TXT",
		Name:    dnsRecordName,
		Content: dnsRecordValue,
	}
}
