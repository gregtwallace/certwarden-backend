//go:build !omitdns01goacme

package dns01goacme

import (
	"certwarden-backend/pkg/datatypes/environment"
	"fmt"

	//nolint:depguard // only using this to fold this external logger into the main logger
	"log/slog"
	"os"

	"go.uber.org/zap/exp/zapslog"

	goacme_dns01 "github.com/go-acme/lego/v5/challenge/dns01"
	goacme_log "github.com/go-acme/lego/v5/log"
	goacme_dns "github.com/go-acme/lego/v5/providers/dns"
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

	// set go-acme log to main zap logger
	goacme_log.SetDefault(slog.New(zapslog.NewHandler(
		service.logger.Desugar().Core(),
		zapslog.WithName("go-acme log"),
		// WithCaller doesn't work as expected (it always returns `log/logger.go`)
		// AddStacktraceAt is unnecessary - already covered by base zap.Logger configuration in main app
	)))

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()

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

		// set global dns01 client to use system name servers
		opts := &goacme_dns01.Options{RecursiveNameservers: dnsServerStrings}
		goacme_dns01.SetDefaultClient(goacme_dns01.NewClient(opts))
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
