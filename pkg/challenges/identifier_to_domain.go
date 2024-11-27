package challenges

import "strings"

// dnsIDValuetoDomain returns the fqdn that should be provisioned to satisfy a given ACME
// DNS Identifier Value. Translation from an identifier to a domain is done via the
// dnsIDtoDomain map.
func (service *Service) dnsIDValuetoDomain(dnsIdentifierValue string) string {
	// split dns identifier into parts
	dnsIDValueSegments := strings.Split(dnsIdentifierValue, ".")

	// check for a match in dnsIDtoDomain map, starting with most specific match
	for index := range dnsIDValueSegments {
		// assemble identifier to check
		identifierValueToCheck := strings.Join(dnsIDValueSegments[index:], ".")

		// look for ID in the map
		domain, exists := service.dnsIDtoDomain.Read(identifierValueToCheck)
		if exists {
			// DONT just return the exact domain; depending on where in the range this is,
			// the beginning of the identifier may need to be prepended
			prependSubDomain := strings.Join(dnsIDValueSegments[:index], ".")
			if prependSubDomain != "" {
				prependSubDomain += "."
			}

			return prependSubDomain + domain
		}
	}

	// No match found, return identifierValue without modification
	return dnsIdentifierValue
}
