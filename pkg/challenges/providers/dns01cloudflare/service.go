package dns01cloudflare

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
}

// Accounts service struct
type Service struct {
	logger        *zap.SugaredLogger
	httpClient    *httpclient.Client
	cloudflareApi *cloudflare.API
	domains       []string
	domainIDs     map[string]string // domain_name[zone_id]
}

// NewService creates a new instance of the Cloudflare provider service. Service
// contains one Cloudflare API instance.
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config or no domains, error
	if cfg == nil || len(cfg.Domains) <= 0 {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// http client for api calls
	service.httpClient = app.GetHttpClient()

	// set supported domains from config
	service.domains = append(service.domains, cfg.Domains...)

	// make map for domains
	service.domainIDs = make(map[string]string)

	// cloudflare api
	err := service.configureCloudflareAPI(cfg)
	if err != nil {
		return nil, err
	}

	// debug log configured domains
	service.logger.Infof("cloudflare instance %s configured domains: %s", service.redactedApiIdentifier(), service.AvailableDomains())

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}
