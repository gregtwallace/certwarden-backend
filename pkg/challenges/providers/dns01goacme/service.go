package dns01goacme

import (
	"certwarden-backend/pkg/acme"
	"errors"

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

// clone of goacme_challenge.Provider (https://github.com/go-acme/lego/blob/v4.35.2/challenge/provider.go#L9)
type provider interface {
	Present(domain, token, keyAuth string) error
	CleanUp(domain, token, keyAuth string) error
}

// provider Service struct
type Service struct {
	logger         *zap.SugaredLogger
	goacmeProvider provider
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
