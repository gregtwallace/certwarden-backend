package dns01cloudflare

import (
	"context"
	"strings"
)

// Provision adds the corresponding DNS record on Cloudflare.
func (service *Service) Provision(resourceName, resourceContent string) error {
	// no need to delete, just handle already exists error (which in theory isn't possible
	// anyway because resourceContent should always change)

	// cloudflare resource
	cfResource, err := service.cloudflareResource(resourceName)
	if err != nil {
		return err
	}

	// create DNS record on cloudflare for the ACME resource
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	_, err = service.cloudflareApi.CreateDNSRecord(ctx, cfResource, cloudflareCreateDNSParams(resourceName, resourceContent))
	if err != nil && !(strings.Contains(err.Error(), "81057") || strings.Contains(err.Error(), "Record already exists")) {
		return err
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(resourceName, resourceContent string) error {
	// cloudflare resource
	cfResource, err := service.cloudflareResource(resourceName)
	if err != nil {
		return err
	}

	// fetch matching record(s) (should only be one)
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	records, _, err := service.cloudflareApi.ListDNSRecords(ctx, cfResource, cloudflareListDNSParams(resourceName, resourceContent))
	if err != nil {
		return err
	}

	// delete all records with the name and content (should only ever be one)
	ctx, cancel = context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	var deleteErr error
	for i := range records {
		err = service.cloudflareApi.DeleteDNSRecord(ctx, cfResource, records[i].ID)
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
