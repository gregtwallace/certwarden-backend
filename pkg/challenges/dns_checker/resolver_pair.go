package dns_checker

import "net"

// dnsIPPair contains a primary and secondary DNS server for
// a given DNS service
type dnsServiceIPPair struct {
	primary   string
	secondary string
}

// dnsResolverPair contains the net.Resolver pair for a specific DNS service
type dnsResolverPair struct {
	primary   *net.Resolver
	secondary *net.Resolver
}

// checkDnsRecord attempts to find the specified record using the dnsResolverPair. It
// first tries the primary dns server and if an error is returned it attempts to use
// the secondary server.
func (rPair dnsResolverPair) checkDnsRecord(fqdn string, recordValue string, recordType dnsRecordType) (exists bool, err error) {
	// try primary
	exists, err = checkDnsRecord(fqdn, recordValue, recordType, rPair.primary)
	// if NO error, return exists
	if err == nil {
		return exists, nil
	}

	// if primary errored, try secondary
	exists, err = checkDnsRecord(fqdn, recordValue, recordType, rPair.secondary)
	// check for secondary error
	if err != nil {
		// return error
		return false, err
	}

	// return if no error
	return exists, nil
}
