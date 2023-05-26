package acme_servers

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"
	"sync"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary acme_servers service component is missing")
	errDuplicateDbValue = errors.New("dupicate custom acme server db_value found")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Acme service struct
type Service struct {
	logger      *zap.SugaredLogger
	httpClient  *httpclient.Client
	acmeServers map[string]acmeServer // [value]acmeServer
	mu          sync.Mutex
}

type Config struct {
	EnableLetsEncrypt *bool              `yaml:"enable_lets_encrypt"`
	EnableGoogleCloud *bool              `yaml:"enable_google_cloud"`
	CustomAcmeServers []acmeServerConfig `yaml:"custom_acme_servers"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
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

	// acme services map
	service.acmeServers = make(map[string]acmeServer)

	// make enabled acme servers array
	acmeServerConfigs := []acmeServerConfig{}
	// LE
	if *cfg.EnableLetsEncrypt {
		acmeServerConfigs = append(acmeServerConfigs, acmeServersLetsEncrypt...)
	}
	// Google Cloud
	if *cfg.EnableGoogleCloud {
		acmeServerConfigs = append(acmeServerConfigs, acmeServersGoogleCloud...)
	}
	// Custom Defined (must convert from config to acmeServer object)
	customDbValues := []string{}
	// add custom servers to list
	for _, customServ := range cfg.CustomAcmeServers {
		dbValue := "custom_" + customServ.Value // prepend custom to the config db_value to avoid conflicts
		// verify none of the custom servers have duplicate values
		for i := range customDbValues {
			if customDbValues[i] == dbValue {
				return nil, errDuplicateDbValue
			}
		}
		// add the value to the dupe check list
		customDbValues = append(customDbValues, dbValue)

		// add server config (with prepended 'custom_' value)
		customServ.Value = dbValue
		acmeServerConfigs = append(acmeServerConfigs, customServ)
	}
	// end make acme servers array

	// populate all of the acme servers services
	// use waitgroup to expedite directory fetching
	var wg sync.WaitGroup
	wgSize := len(acmeServerConfigs)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// for each CA in config, configure ACME
	for i := range acmeServerConfigs {
		go func(asc acmeServerConfig) {
			// done after func
			defer wg.Done()

			// make service
			var err error
			as := acmeServer{
				acmeServerConfig: asc,
			}
			as.service, err = acme.NewService(app, as.DirUri)
			wgErrors <- err

			// don't directly assign to map so dir fetching can occur simultaneously
			service.mu.Lock()
			service.acmeServers[as.Value] = as
			service.mu.Unlock()
		}(acmeServerConfigs[i])
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			service.logger.Errorf("failed to configure app acme service(s) (%s)", err)
			return nil, err
		}
	}
	// end acme services

	return service, nil
}
