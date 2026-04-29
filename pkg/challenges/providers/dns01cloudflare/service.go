package dns01cloudflare

import (
	"certwarden-backend/pkg/acme"
	"context"
	"errors"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v6"
	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetHttpClient() *http.Client
}

// provider Service struct
type Service struct {
	logger           *zap.SugaredLogger
	shutdownContext  context.Context
	httpClient       *http.Client
	cloudflareClient *cloudflare.Client
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// NewService creates a new instance of the Cloudflare provider service. Service
// contains one Cloudflare API instance.
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

	// shutdown ctx
	service.shutdownContext = app.GetShutdownContext()

	// http client for api calls
	service.httpClient = app.GetHttpClient()

	// cloudflare client
	err := service.configureCloudflareClient(cfg)
	if err != nil {
		return nil, err
	}

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
