//go:build !omitdns01goacme

package dns01goacme

import (
	"certwarden-backend/pkg/datatypes/environment"
	"fmt"
	"os"

	goacme_dns01 "github.com/go-acme/lego/v4/challenge/dns01"
	goacme_dns "github.com/go-acme/lego/v4/providers/dns"
)

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	if cfg == nil {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// set environment
	envParams, invalidParams := environment.NewParams(cfg.Environment)
	envMap := envParams.KeyValMap()
	if len(invalidParams) > 0 {
		service.logger.Errorf("dns-01 go-acme some environment param(s) invalid and won't be used (%s)", invalidParams)
	}
	for key, val := range envMap {
		err := os.Setenv(key, val)
		if err != nil {
			return nil, fmt.Errorf("go-acme failed to set environment variable (%s)", err)
		}
	}

	// go-acme annoyingly does dns lookups - try to deduce the system dns servers
	// and use them (if none found, no-op, which go-acme will use its default)
	dnsServers := GetDNSServers()
	if len(dnsServers) > 0 {
		dnsServerStrings := []string{}
		for _, dnsServ := range dnsServers {
			dnsServerStrings = append(dnsServerStrings, dnsServ.String())
		}
		// note: AddRecursiveNameservers returns a func that sets go-acme's 'global' dns servers;
		// call this func (use nil since the func doesn't actually use the challenge) to set the dns servers
		goacme_dns01.AddRecursiveNameservers(dnsServerStrings)(nil)
	}

	// make go acme provider
	var err error
	service.goacmeProvider, err = goacme_dns.NewDNSChallengeProviderByName(cfg.DnsProviderName)
	if err != nil {
		return nil, fmt.Errorf("failed to configure go-acme dns provider (%s)", err)
	}

	// clear environment (only needed during creation of the go-acme provider)
	for key := range envMap {
		err := os.Unsetenv(key)
		if err != nil {
			return nil, fmt.Errorf("go-acme failed to clear environment variable (%s)", err)
		}
	}

	return service, nil
}
