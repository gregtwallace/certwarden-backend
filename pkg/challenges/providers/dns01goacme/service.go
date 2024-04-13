package dns01goacme

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/datatypes/environment"
	"errors"
	"fmt"
	"os"

	goacme_challenge "github.com/go-acme/lego/v4/challenge"
	goacme_dns01 "github.com/go-acme/lego/v4/challenge/dns01"
	goacme_dns "github.com/go-acme/lego/v4/providers/dns"
	"go.uber.org/zap"
)

// WARNING: Trying to make multiple providers of this type at once can cause problems if
// same dns provider is being used (environment variables could overwrite). Don't do that!

var (
	errServiceComponent = errors.New("necessary dns-01 go-acme component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Configuration options; documentation about how to configure is located at:
// https://go-acme.github.io/lego/dns/
type Config struct {
	// Available in docs as "CLI flag name" or "Code"
	DnsProviderName string `json:"dns_provider_name"`
	// available options listed for each provider in go-acme docs
	Environment []string `yaml:"environment" json:"environment"`
}

// provider Service struct
type Service struct {
	logger         *zap.SugaredLogger
	goacmeProvider goacme_challenge.Provider
}

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

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) error {
	// if no config, error
	if cfg == nil {
		return errServiceComponent
	}

	// don't need to do anything with "old" Service, just set a new one
	newServ, err := NewService(app, cfg)
	if err != nil {
		return err
	}

	// set content of old pointer so anything with the pointer calls the
	// updated service
	*service = *newServ

	return nil
}
