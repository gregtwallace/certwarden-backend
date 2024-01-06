package dns01cloudflare

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"

	"github.com/cloudflare/cloudflare-go"
)

// Provision adds the corresponding DNS record on Cloudflare.
func (service *Service) Provision(domain, _, keyAuth string) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// cloudflare resource
	cfResource, err := service.cloudflareResource(dnsRecordName)
	if err != nil {
		return err
	}

	// create DNS record on cloudflare for the ACME resource
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	_, err = service.cloudflareApi.CreateDNSRecord(ctx, cfResource, cloudflareCreateDNSParams(dnsRecordName, dnsRecordValue))
	cfReqErr := new(cloudflare.RequestError)
	// return err if not a CF RequestError, or if it is CF Req Error, but does NOT contain code 81057 (which is "Record already exists")
	if err != nil && (!errors.As(err, &cfReqErr) || !cfReqErr.InternalErrorCodeIs(81057)) {
		return err
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(domain, _, keyAuth string) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// cloudflare resource
	cfResource, err := service.cloudflareResource(dnsRecordName)
	if err != nil {
		return err
	}

	// fetch matching record(s) (should only be one)
	ctx, cancel := context.WithTimeout(context.Background(), apiCallTimeout)
	defer cancel()

	records, _, err := service.cloudflareApi.ListDNSRecords(ctx, cfResource, cloudflareListDNSParams(dnsRecordName, dnsRecordValue))
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
