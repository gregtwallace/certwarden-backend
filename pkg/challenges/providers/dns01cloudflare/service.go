package dns01cloudflare

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"

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

// provider Service struct
type Service struct {
	logger        *zap.SugaredLogger
	httpClient    *httpclient.Client
	cloudflareApi *cloudflare.API
	domains       []string
	domainIDs     map[string]string // domain_name[zone_id]
}

// Stop/Start is not needed for this provider. Nothing needs to be stopped or started.
func (service *Service) Stop() error  { return nil }
func (service *Service) Start() error { return nil }

// NewService creates a new instance of the Cloudflare provider service. Service
// contains one Cloudflare API instance.
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config or no domains, error
	if cfg == nil || len(cfg.Doms) <= 0 {
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
	service.domains = append(service.domains, cfg.Doms...)

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

// redactedIdentifier selects either the APIKey, APIUserServiceKey, or APIToken
// (depending on which is in use for the API instance) and then redacts it to return
// the first and last characters of the key separated with asterisks. This is useful
// for logging issues without saving the full credential to logs.
func (service *Service) redactedApiIdentifier() string {
	identifier := ""

	// select whichever is present
	if len(service.cloudflareApi.APIToken) > 0 {
		identifier = service.cloudflareApi.APIToken
	} else if len(service.cloudflareApi.APIKey) > 0 {
		identifier = service.cloudflareApi.APIKey
	} else if len(service.cloudflareApi.APIUserServiceKey) > 0 {
		identifier = service.cloudflareApi.APIUserServiceKey
	} else {
		// none present, return unknown
		return "unknown"
	}

	// return redacted
	return output.RedactString(identifier)
}
