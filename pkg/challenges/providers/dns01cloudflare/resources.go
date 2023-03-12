package dns01cloudflare

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/challenges/dns_checker"
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

// Provision adds the resource to the internal tracking map and provisions
// the corresponding DNS record on Cloudflare.
func (service *Service) Provision(resourceName string, resourceContent string) error {
	// add to internal map
	exists, existingContent := service.dnsRecords.Add(resourceName, resourceContent)
	// if already exists, but content is different, error
	if exists && existingContent != resourceContent {
		return fmt.Errorf("dns-01 (cloudflare) can't add resource (%s), already exists "+
			"and content does not match", resourceName)
	}

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

	// check for propagation
	propagated, err := service.dnsChecker.CheckTXTWithRetry(resourceName, resourceContent, 10)
	if err != nil {
		service.logger.Error(err)
		return err
	}

	// if failed to propagate
	if !propagated {
		return dns_checker.ErrDnsRecordNotFound
	}

	return nil
}

// Deprovision removes the resource from the internal tracking map and deletes
// the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(resourceName string, resourceContent string) error {
	// remove from internal map
	err := service.dnsRecords.Delete(resourceName)
	if err != nil {
		service.logger.Errorf("dns-01 (cloudflare) could not remove resource (%s) from "+
			"internal map", resourceName)
		// do not return
	}

	// get the relevant zone from known list
	zone, err := service.getResourceZone(resourceName)
	if err != nil {
		return ErrDomainNotConfigured
	}

	// remove old DNS record
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
	// remove old DNS record - END

	return nil
}
