package dns01acmedns

import (
	"errors"
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

// Accounts service struct
type Service struct {
	logger           *zap.SugaredLogger
	httpClient       *httpclient.Client
	acmeDnsAddress   string
	acmeDnsResources []acmeDnsResource
}

// Configuration options
type Config struct {
	Enable      *bool             `yaml:"enable"`
	HostAddress *string           `yaml:"acme_dns_address"`
	Resources   []acmeDnsResource `yaml:"resources"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if disabled, return nil and no error
	if !*cfg.Enable {
		return nil, nil
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

	// acme-dns host address
	service.acmeDnsAddress = *cfg.HostAddress

	// acme-dns resources that will be updated
	service.acmeDnsResources = cfg.Resources

	return service, nil
}
