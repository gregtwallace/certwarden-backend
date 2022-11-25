package dns01cloudflare

import (
	"context"
	"errors"
	"fmt"
	"time"
)

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

	// get ZoneID
	zoneID, err := service.getZoneID(resourceName)
	if err != nil {
		return err
	}

	// remove old DNS record, if one exists
	err = service.deleteDNSRecord(resourceName, zoneID)
	if err != nil {
		return err
	}

	// create DNS record on cloudflare for the ACME resource
	_, err = service.cloudflareApi.CreateDNSRecord(context.Background(), zoneID, newAcmeRecord(resourceName, resourceContent))
	if err != nil {
		return err
	}

	// confirm dns record has propagated before returning result
	// retry propagation check a few times before failing
	maxTries := 5
	for i := 1; i <= maxTries; i++ {
		// sleep at start to allow some propagation
		time.Sleep(time.Duration(i) * 15 * time.Second)

		// check for propagation
		propagated, err := service.dnsChecker.CheckTXT(resourceName, resourceContent)
		// if error, log error but still retry
		if err != nil {
			service.logger.Error(err)
		}

		// if propagated, done & success
		if propagated {
			return nil
		}
	}

	return errors.New("failed to propagate dns record")
}

func (service *Service) Deprovision(resourceName string) error {
	// remove from internal map
	err := service.dnsRecords.Delete(resourceName)
	if err != nil {
		service.logger.Errorf("dns-01 (cloudflare) could not remove resource (%s) from "+
			"internal map", resourceName)
		// do not return
	}

	// get ZoneID
	zoneID, err := service.getZoneID(resourceName)
	if err != nil {
		return err
	}

	// remove old DNS record
	err = service.deleteDNSRecord(resourceName, zoneID)
	if err != nil {
		return err
	}

	return nil
}
