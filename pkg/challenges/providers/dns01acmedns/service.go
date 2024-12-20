package dns01acmedns

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/httpclient"
	"errors"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 acme-dns component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
}

// provider Service struct
type Service struct {
	logger           *zap.SugaredLogger
	httpClient       *httpclient.Client
	acmeDnsAddress   string
	acmeDnsResources []acmeDnsResource
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// check config
	err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	service := new(Service)

	// http client
	service.httpClient = app.GetHttpClient()
	if service.httpClient == nil {
		return nil, errServiceComponent
	}

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// acme-dns host address
	service.acmeDnsAddress = cfg.HostAddress

	// acme-dns resources that will be updated
	service.acmeDnsResources = cfg.Resources

	return service, nil
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
