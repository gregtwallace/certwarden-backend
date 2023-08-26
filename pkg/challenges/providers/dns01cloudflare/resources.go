package dns01cloudflare

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

var ErrDomainNotConfigured = errors.New("dns01cloudflare domain name not configured (restart lego if zone was just added to an account)")

// AvailableDomains returns all of the domains that this instance of Cloudflare can
// provision records for. This is separate from all possible supported domains as
// challenges may not actually want to use this provider for all possible domains.
func (service *Service) AvailableDomains() []string {
	return service.domains
}

// getZoneId returns the zone ID for a specified resourceName. If a matching
// zone isn't found, an error is returned.
func (service *Service) getZoneId(resourceName string) (string, error) {
	// look for domain suffix
	for domain := range service.domainIDs {
		if strings.HasSuffix(resourceName, domain) {
			return service.domainIDs[domain], nil
		}
	}

	return "", fmt.Errorf("could not find domain supporting resource name %s", resourceName)
}

// acmeRecord returns the cloudflare dns record for a given acme resource
// name and content
func newAcmeRecord(resourceName, resourceContent string) cloudflare.DNSRecord {
	return cloudflare.DNSRecord{
		Type:      "TXT",
		Name:      resourceName,
		Content:   resourceContent,
		TTL:       60,
		Proxiable: false,
	}
}

// Provision adds the corresponding DNS record on Cloudflare.
func (service *Service) Provision(resourceName, resourceContent string) error {
	// no need to delete, just handle already exists error (which in theory isn't possible
	// anyway because resourceContent should always change)

	// zoneID
	zoneID, err := service.getZoneId(resourceName)
	if err != nil {
		return err
	}

	// create DNS record on cloudflare for the ACME resource
	_, err = service.cloudflareApi.CreateDNSRecord(context.Background(), zoneID, newAcmeRecord(resourceName, resourceContent))
	if err != nil && !(strings.Contains(err.Error(), "81057") || strings.Contains(err.Error(), "Record already exists")) {
		return err
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(resourceName, resourceContent string) error {
	// zoneID
	zoneID, err := service.getZoneId(resourceName)
	if err != nil {
		return err
	}

	// fetch matching record(s) (should only be one)
	records, err := service.cloudflareApi.DNSRecords(context.Background(), zoneID, cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    resourceName,
		Content: resourceContent,
	})
	if err != nil {
		return err
	}

	// delete all records with the name and content (should only ever be one)
	var deleteErr error
	for i := range records {
		err = service.cloudflareApi.DeleteDNSRecord(context.Background(), zoneID, records[i].ID)
		if err != nil {
			deleteErr = err
			service.logger.Error(err)
		}
	}

	if deleteErr != nil {
		return err
	}

	return nil
}
