package dns01cloudflare

import (
	"certwarden-backend/pkg/acme"
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
)

// Provision adds the corresponding DNS record on Cloudflare.
func (service *Service) Provision(domain string, _ string, keyAuth acme.KeyAuth) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// get zone
	zoneID, err := service.getZoneID(dnsRecordName)
	if err != nil {
		return fmt.Errorf("dns01cloudflare: failed to get zone id for %s (%s)", dnsRecordName, err)
	}

	// create DNS record on cloudflare for the ACME resource
	ctx, cancel := context.WithTimeout(service.shutdownContext, apiCallTimeout)
	defer cancel()

	_, err = service.cloudflareClient.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   cloudflareCreateDNSParams(dnsRecordName, dnsRecordValue),
	})
	if err != nil {
		// try to check cloudflare error
		cfErr, ok := err.(*cloudflare.Error)
		if !ok {
			return fmt.Errorf("dns01cloudflare: failed to create dns record %s: %s (%s)", dnsRecordName, dnsRecordValue, err)
		}

		// record exists error (81057 or 81058) is fine
		alreadyExistsError := false
		for i := range cfErr.Errors {
			if cfErr.Errors[i].Code == 81057 || cfErr.Errors[i].Code == 81058 {
				alreadyExistsError = true
				break
			}
		}
		if !alreadyExistsError {
			return fmt.Errorf("dns01cloudflare: failed to create dns record %s: %s (%s)", dnsRecordName, dnsRecordValue, err)
		}
	}

	return nil
}

// Deprovision deletes the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(domain string, _ string, keyAuth acme.KeyAuth) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// get zone
	zoneID, err := service.getZoneID(dnsRecordName)
	if err != nil {
		return fmt.Errorf("dns01cloudflare: failed to get zone id for %s (%s)", dnsRecordName, err)
	}

	// fetch matching record(s) (should only be one)
	ctx, cancel := context.WithTimeout(service.shutdownContext, apiCallTimeout)
	defer cancel()

	resultPage, err := service.cloudflareClient.DNS.Records.List(ctx, cloudflareListDNSParams(dnsRecordName, dnsRecordValue, zoneID))
	if err != nil {
		return fmt.Errorf("dns01cloudflare: failed to delete %s: %s (couldn't list dns records) (%s)", dnsRecordName, dnsRecordValue, err)
	}

	// don't bother checking addl pages, if there are over 100 records somehow, that's beyond the help of this application
	dnsRecordIDs := []string{}
	for i := range resultPage.Result {
		dnsRecordIDs = append(dnsRecordIDs, resultPage.Result[i].ID)
	}

	// delete all records with the name and content (should only ever be one)
	anyDeleteErr := false
	for _, recordID := range dnsRecordIDs {
		ctx, cancel = context.WithTimeout(service.shutdownContext, apiCallTimeout)
		defer cancel()

		_, err = service.cloudflareClient.DNS.Records.Delete(
			ctx,
			recordID,
			cloudflareDeleteDNSParams(zoneID),
		)
		if err != nil {
			anyDeleteErr = true
			service.logger.Errorf("dns01cloudflare: failed to delete %s: %s (record ID: %s) (%s)", dnsRecordName, dnsRecordValue, recordID, err)
		}
	}

	if anyDeleteErr {
		return fmt.Errorf("dns01cloudflare: failed to delete dns record(s) during cleanup step of %s: %s", dnsRecordName, dnsRecordValue)
	}

	return nil
}
