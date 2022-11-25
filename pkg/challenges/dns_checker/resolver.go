package dns_checker

import (
	"context"
	"net"
	"time"
)

// timeoutSeconds is the DNS dialer timeout (in seconds)
const timeoutSeconds = 5

// makeResolvers generates all of the resolver pairs for a slice
// of DNS Service IP Pairs
func makeResolvers(dnsServices []dnsServiceIPPair) []dnsResolverPair {
	dnsResolverPairs := []dnsResolverPair{}

	// add each service pair to the resolver pairs
	for i := range dnsServices {
		servicePair := dnsResolverPair{
			primary:   makeResolver(dnsServices[i].primary),
			secondary: makeResolver(dnsServices[i].secondary),
		}

		dnsResolverPairs = append(dnsResolverPairs, servicePair)
	}

	return dnsResolverPairs
}

// makeResolver creates a net.Resolver to resolve DNS queries using
// the specified DNS server IP.
func makeResolver(ipAddress string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: timeoutSeconds * time.Second,
			}
			return d.DialContext(ctx, network, ipAddress+":53")
		},
	}
}
