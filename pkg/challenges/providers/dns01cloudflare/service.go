package dns01cloudflare

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/datatypes"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	logger        *zap.SugaredLogger
	dnsChecker    *dns_checker.Service
	cloudflareApi *cloudflare.API
	dnsRecords    *datatypes.SafeMap
}

// Configuration options
type Config struct {
	Enable        *bool  `yaml:"enabled"`
	AccountEmail  string `yaml:"account_email"`
	AccountApiKey string `yaml:"account_api_key"`
}

// NewService creates a new service
func NewService(app App, config *Config, dnsChecker *dns_checker.Service) (*Service, error) {
	// if disabled, return nil and no error
	if !*config.Enable {
		return nil, nil
	}

	service := new(Service)
	var err error

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// dns checker
	service.dnsChecker = dnsChecker

	// cloudflare api
	service.cloudflareApi, err = cloudflare.New(config.AccountApiKey, config.AccountEmail)
	if err != nil {
		return nil, err
	}

	// Verify the account credentials actually work
	_, err = service.cloudflareApi.UserDetails(context.Background())
	if err != nil {
		return nil, err
	}

	// map to hold current dnsRecords
	service.dnsRecords = datatypes.NewSafeMap()

	return service, nil
}
