package dns01cloudflare

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

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

// getZoneID returns the ZoneID for a specific resourceName
func (service *Service) getZoneID(resourceName string) (zoneID string, err error) {
	// determine the resource TLD (i.e. the ZoneID Name)
	domainParts := strings.Split(resourceName, ".")
	zoneName := domainParts[len(domainParts)-2] + "." + domainParts[len(domainParts)-1]

	// get ZoneID
	zoneID, err = service.cloudflareApi.ZoneIDByName(zoneName)
	if err != nil {
		return "", err
	}

	return zoneID, nil
}

// deleteDNSRecord deletes the resourceName's record with the specified content from
// the specified zoneID, if the record exists. If it does not exist, it does nothing.
func (service *Service) deleteDNSRecord(resourceName string, resourceContent string, zoneID string) (err error) {
	// fetch matching record(s) (should only be one)
	records, err := service.cloudflareApi.DNSRecords(context.Background(), zoneID, cloudflare.DNSRecord{
		Type:    "TXT",
		Name:    resourceName,
		Content: resourceContent,
	})
	if err != nil {
		return err
	}

	// if doesn't exist, done
	if len(records) == 0 {
		return nil
	}

	// does exist, delete all records with the name and content (should only ever be one)
	for i := range records {
		err = service.cloudflareApi.DeleteDNSRecord(context.Background(), zoneID, records[i].ID)
		if err != nil {
			service.logger.Error(err)
		}
	}

	if err != nil {
		return err
	}

	return nil
}
