package challenges

import (
	"certwarden-backend/pkg/challenges/providers"
	"certwarden-backend/pkg/datatypes/safemap"
	"certwarden-backend/pkg/output"
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// use a conservative rate limit to avoid any issues
const (
	limitEventsPerSecond = 2
	limitBurstAtMost     = 2
)

var (
	errServiceComponent = errors.New("challenges: necessary service component is missing")
)

// App interface is for connecting to the main app
type application interface {
	GetConfigFilenameWithPath() string
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
	GetOutputter() *output.Service

	// for providers
	GetHttpClient() *http.Client
}

// Config holds all of the challenge config
type Config struct {
	ProviderConfigs providers.Config  `yaml:"providers"`
	DNSIDtoDomain   map[string]string `yaml:"domain_aliases"`
}

// service struct
type Service struct {
	app                    application
	logger                 *zap.SugaredLogger
	shutdownContext        context.Context
	shutdownWaitgroup      *sync.WaitGroup
	output                 *output.Service
	configFile             string
	DNSIdentifierProviders *providers.Manager
	dnsIDtoDomain          *safemap.SafeMap[string] // DNSIdentifierValue[Domain]
	apiRateLimiter         *rate.Limiter
}

// NewService creates a new service
func NewService(app application, cfg *Config) (service *Service, err error) {
	service = new(Service)

	// save app pointer for use later in reconfiguring providers
	service.app = app

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output
	service.output = app.GetOutputter()

	// config file path (for writing)
	service.configFile = app.GetConfigFilenameWithPath()

	// shutdown context & wg
	service.shutdownContext = app.GetShutdownContext()
	service.shutdownWaitgroup = app.GetShutdownWaitGroup()

	// configure challenge providers
	service.DNSIdentifierProviders, err = providers.MakeManager(app, cfg.ProviderConfigs)
	if err != nil {
		service.logger.Errorf("challenges: failed to configure challenge provider(s) (%s)", err)
		return nil, err
	}

	// api management rate limiter
	service.apiRateLimiter = rate.NewLimiter(rate.Every(time.Second/limitEventsPerSecond), limitBurstAtMost)

	// make DNS Identifier -> domain map (from config value)
	service.dnsIDtoDomain = safemap.NewSafeMapFrom(cfg.DNSIDtoDomain)

	return service, nil
}
