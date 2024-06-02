package dns_checker

import (
	"context"
	"errors"
	"net"
	"time"
)

// timeoutSeconds is the DNS dialer timeout (in seconds)
const timeoutSeconds = 5

var (
	errBlankIP = errors.New("dns_checker: can't create resolver, ip address is blank")
)

// makeResolvers generates all of the resolver pairs for a slice
// of DNS Service IP Pairs
func makeResolvers(dnsServices []DnsServiceIPPair) ([]dnsResolverPair, error) {
	// add each service pair to the resolver pairs
	dnsResolverPairs := []dnsResolverPair{}
	for i := range dnsServices {
		// make primary
		primaryR, err := makeResolver(dnsServices[i].Primary)
		if err != nil {
			return nil, err
		}

		// make secondary (blank is okay, just exclude it)
		secondaryR, err := makeResolver(dnsServices[i].Secondary)
		if err != nil && !errors.Is(err, errBlankIP) {
			return nil, err
		}

		// make pair
		servicePair := dnsResolverPair{
			primary:   primaryR,
			secondary: secondaryR,
		}

		// append to list of pairs
		dnsResolverPairs = append(dnsResolverPairs, servicePair)
	}

	return dnsResolverPairs, nil
}

// makeResolver creates a net.Resolver to resolve DNS queries using
// the specified DNS server IP.
func makeResolver(ipAddress string) (*net.Resolver, error) {
	if ipAddress == "" {
		return nil, errBlankIP
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, ipAddress+":53")
		},
	}

	// make sure the dns resolver actually works
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	_, err := r.LookupIP(ctx, "ip", "google.com")
	if err != nil {
		return nil, err
	}

	return r, nil
}
