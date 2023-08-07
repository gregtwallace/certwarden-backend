package dns01cloudflare

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

var ErrDomainNotConfigured = errors.New("dns01cloudflare domain name not configured (restart lego if zone was just added to an account)")

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
func (service *Service) Provision(resourceName string, resourceContent string) error {
	// get the relevant zone from known list
	zone, err := service.getResourceZone(resourceName)
	if err != nil {
		return ErrDomainNotConfigured
	}

	// no need to delete, just handle already exists error (which in theory isn't possible
	// anyway because resourceContent should always change)

	// create DNS record on cloudflare for the ACME resource
	_, err = zone.api.CreateDNSRecord(context.Background(), zone.id, newAcmeRecord(resourceName, resourceContent))
	if err != nil && !(strings.Contains(err.Error(), "81057") || strings.Contains(err.Error(), "Record already exists")) {
		return err
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(resourceName string, resourceContent string) error {
	// get the relevant zone from known list
	zone, err := service.getResourceZone(resourceName)
	if err != nil {
		return ErrDomainNotConfigured
	}

	// fetch matching record(s) (should only be one)
	records, err := zone.api.DNSRecords(context.Background(), zone.id, cloudflare.DNSRecord{
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
		err = zone.api.DeleteDNSRecord(context.Background(), zone.id, records[i].ID)
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
