package dns01acmesh

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/datatypes/environment"
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary dns-01 acme.sh component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// provider Service struct
type Service struct {
	logger            *zap.SugaredLogger
	shellPath         string
	acmeShPath        string
	dnsHook           string
	environmentParams *environment.Params
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// Configuration options
type Config struct {
	AcmeShPath  string   `yaml:"acme_sh_path" json:"acme_sh_path"`
	Environment []string `yaml:"environment" json:"environment"`
	DnsHook     string   `yaml:"dns_hook" json:"dns_hook"`
}

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
