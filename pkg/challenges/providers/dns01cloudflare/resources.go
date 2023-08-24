package dns01cloudflare

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

var ErrDomainNotConfigured = errors.New("dns01cloudflare domain name not configured (restart lego if zone was just added to an account)")

// AvailableDomains returns all of the domains that this instance of Cloudflare can
// provision records for.
func (service *Service) AvailableDomains() []string {
	domainList := []string{}
	for domainName := range service.domainIDs {
		domainList = append(domainList, domainName)
	}

	return domainList
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
func (service *Service) Provision(domainName, resourceName, resourceContent string) error {
	// no need to delete, just handle already exists error (which in theory isn't possible
	// anyway because resourceContent should always change)

	// create DNS record on cloudflare for the ACME resource
	_, err := service.cloudflareApi.CreateDNSRecord(context.Background(), service.domainIDs[domainName], newAcmeRecord(resourceName, resourceContent))
	if err != nil && !(strings.Contains(err.Error(), "81057") || strings.Contains(err.Error(), "Record already exists")) {
		return err
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(domainName, resourceName, resourceContent string) error {
	// zoneID
	zoneID := service.domainIDs[domainName]

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
