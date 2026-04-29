package dns01cloudflare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

// getZoneID returns the cloudflare zone id for the domain or subdomain specified; if the API call
// fails or the zone isn't found, an error is returned
func (service *Service) getZoneID(dnsRecordName string) (string, error) {
	domainParts := strings.Split(dnsRecordName, ".")
	checkFor := ""
	zoneID := ""

	for i := range len(domainParts) {
		if i != 0 {
			checkFor = "." + checkFor
		}
		checkFor = domainParts[len(domainParts)-i-1] + checkFor

		// never check the first piece
		if i == 0 {
			continue
		}

		ctx, cancel := context.WithTimeout(service.shutdownContext, apiCallTimeout)
		defer cancel()
		resp, err := service.cloudflareClient.Zones.List(ctx, zones.ZoneListParams{
			Name: cloudflare.F("equal:" + checkFor),
		})
		if err != nil {
			service.logger.Errorf("dns01cloudflare: list zones api call failed (%s)", err)
			continue
		}

		if len(resp.Result) > 0 {
			zoneID = resp.Result[0].ID
			break
		}
	}

	if zoneID == "" {
		return "", fmt.Errorf("could not find cloudflare zone for %s", dnsRecordName)
	}

	return zoneID, nil
}

// cloudflareCreateDNSParams returns the cloudflare create dns record params for a given
// acme resource name and content
func cloudflareCreateDNSParams(dnsRecordName, dnsRecordValue string) dns.TXTRecordParam {
	return dns.TXTRecordParam{
		Name:    cloudflare.F(dnsRecordName),
		Content: cloudflare.F("\"" + dnsRecordValue + "\""),
		Type:    cloudflare.F(dns.TXTRecordTypeTXT),

		// specific to create
		TTL:     cloudflare.F(dns.TTL(60)), // 60 seconds
		Proxied: cloudflare.Bool(false),
		Comment: cloudflare.F(fmt.Sprintf("created by Cert Warden on %s", time.Now().Format("Mon Jan 2 15:04:05 MST 2006"))),
	}
}

// cloudflareListDNSParams returns the cloudflare list dns records params for a given
// acme resource name and content
func cloudflareListDNSParams(dnsRecordName, dnsRecordValue, zoneID string) dns.RecordListParams {
	return dns.RecordListParams{
		ZoneID:  cloudflare.F(zoneID),
		Name:    cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(dnsRecordName)}),
		Content: cloudflare.F(dns.RecordListParamsContent{Exact: cloudflare.F("\"" + dnsRecordValue + "\"")}),
		Type:    cloudflare.F(dns.RecordListParamsTypeTXT),
	}
}

// cloudflareDeleteDNSParams returns the cloudflare delete dns records params for a zoneID
func cloudflareDeleteDNSParams(zoneID string) dns.RecordDeleteParams {
	return dns.RecordDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	}
}
