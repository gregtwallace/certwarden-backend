package acme

import (
	"crypto/sha256"
)

// ValidationResourceDns01 returns the dnsRecord name and value to provision
// in response to a Dns01 challenge for a given domain and keyAuth
func ValidationResourceDns01(domain, keyAuth string) (dnsRecordName, dnsRecordValue string) {
	// dns record name is just the domain prepended with the special acme prefix
	dnsRecordName = "_acme-challenge." + domain

	// dns record value is the base64 encoded sha256 of key authorization
	// calculate digest
	keyAuthDigest := sha256.Sum256([]byte(keyAuth))

	// encode it
	dnsRecordValue = encodeString(keyAuthDigest[:])

	return dnsRecordName, dnsRecordValue
}
