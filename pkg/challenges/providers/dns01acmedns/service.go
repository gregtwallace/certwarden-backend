package dns01acmedns

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"

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
	domains          []string
	acmeDnsAddress   string
	acmeDnsResources []acmeDnsResource
}

// Stop/Start is not needed for this provider. Nothing needs to be stopped or started.
func (service *Service) Stop() error  { return nil }
func (service *Service) Start() error { return nil }

// Configuration options
type Config struct {
	Doms        []string          `yaml:"domains" json:"domains"`
	HostAddress *string           `yaml:"acme_dns_address" json:"acme_dns_address"`
	Resources   []acmeDnsResource `yaml:"resources" json:"resources"`
}

// Domains returns all of the domains specified in the Config
func (cfg *Config) Domains() []string {
	return cfg.Doms
}

// NewService creates a new service
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

	// http client
	service.httpClient = app.GetHttpClient()
	if service.httpClient == nil {
		return nil, errServiceComponent
	}

	// set supported domains from config
	service.domains = cfg.Doms

	// acme-dns host address
	service.acmeDnsAddress = *cfg.HostAddress

	// acme-dns resources that will be updated
	service.acmeDnsResources = cfg.Resources

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}
