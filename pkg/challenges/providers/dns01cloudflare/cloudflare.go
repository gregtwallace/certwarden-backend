package dns01cloudflare

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

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

// deleteDNSRecord deletes the resourceName's record from the specified
// zoneID, if the resource already has a DNS record. Otherwise, it does
// nothing.
func (service *Service) deleteDNSRecord(resourceName string, zoneID string) (err error) {
	records, err := service.cloudflareApi.DNSRecords(context.Background(), zoneID, cloudflare.DNSRecord{
		Type: "TXT",
		Name: resourceName,
	})
	if err != nil {
		return err
	}

	// if doesn't exist, done
	if len(records) == 0 {
		return nil
	}

	// does exist, delete it
	err = service.cloudflareApi.DeleteDNSRecord(context.Background(), zoneID, records[0].ID)
	if err != nil {
		return err
	}

	return nil
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

// findAcmeRecord searches for an ACME record with the given resource name
// and zoneID and returns the record's ID if it exists. If it does not exist,
// an empty string is returned instead.
// func (service *Service) findAcmeRecord(resourceName string, zoneID string) (recordID string, err error) {
// 	records, err := service.cloudflareApi.DNSRecords(context.Background(), zoneID, cloudflare.DNSRecord{
// 		Type: "TXT",
// 		Name: resourceName,
// 	})
// 	if err != nil {
// 		return "", err
// 	}

// 	return records[0].ID, nil
// }
