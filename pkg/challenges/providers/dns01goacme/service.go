package dns01goacme

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"os"
	"strings"

	goacme_challenge "github.com/go-acme/lego/v4/challenge"
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
	Environment output.RedactedEnvironmentParams `yaml:"environment" json:"environment"`
}

// provider Service struct
type Service struct {
	logger          *zap.SugaredLogger
	environmentVars []string
	goacmeProvider  goacme_challenge.Provider
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

	// environment vars
	service.environmentVars = cfg.Environment.Unredacted()

	// set environment vars
	for _, envVar := range service.environmentVars {
		envVarPieces := strings.Split(envVar, "=")
		if len(envVarPieces) != 2 {
			return nil, errors.New("go-acme environment variable is improperly formatted, should be name=value")
		}
		err := os.Setenv(envVarPieces[0], envVarPieces[1])
		if err != nil {
			return nil, fmt.Errorf("go-acme failed to set environment variable (%s)", err)
		}
	}

	// make go acme provider
	var err error
	service.goacmeProvider, err = goacme_dns.NewDNSChallengeProviderByName(cfg.DnsProviderName)
	if err != nil {
		return nil, fmt.Errorf("failed to configure go-acme dns provider (%s)", err)
	}

	// clear environment vars (only needed during creation of go-acme provider)
	for _, envVar := range service.environmentVars {
		envVarPieces := strings.Split(envVar, "=")
		// skip len check, already done during creation
		err := os.Unsetenv(envVarPieces[0])
		if err != nil {
			// don't exit, just log
			service.logger.Errorf("go-acme failed to unset environment variable (%s)", err)
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

	// try to fix redacted vals from client
	if cfg.Environment != nil {
		cfg.Environment.TryUnredact(service.environmentVars)
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
