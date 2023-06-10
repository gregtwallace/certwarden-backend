package acme_servers

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/pagination_sort"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary acme_servers service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetAcmeServerStorage() Storage
	GetHttpClient() *httpclient.Client
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Storage interface for storage functions
type Storage interface {
	GetAllAcmeServers(q pagination_sort.Query) ([]Server, int, error)
}

// Acme service struct
type Service struct {
	logger      *zap.SugaredLogger
	storage     Storage
	httpClient  *httpclient.Client
	acmeServers map[int]*acme.Service // [id]acmeServer
	mu          sync.Mutex
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetAcmeServerStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	// http client
	service.httpClient = app.GetHttpClient()
	if service.httpClient == nil {
		return nil, errServiceComponent
	}

	// acme services map
	service.acmeServers = make(map[int]*acme.Service)

	// get server list from storage
	servers, _, err := service.storage.GetAllAcmeServers(pagination_sort.QueryAll)
	if err != nil {
		return nil, err
	}

	// populate all of the acme servers services
	// use waitgroup to expedite directory fetching
	var wg sync.WaitGroup
	wgSize := len(servers)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// for each server, configure ACME service
	for i := range servers {
		go func(serv Server) {
			// done after func
			defer wg.Done()

			// make service
			acmeService, err := acme.NewService(app, serv.DirectoryURL)
			wgErrors <- err

			// don't directly assign to map so dir fetching can occur simultaneously
			service.mu.Lock()
			service.acmeServers[serv.ID] = acmeService
			service.mu.Unlock()
		}(servers[i])
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			service.logger.Errorf("failed to configure app acme server(s) (%s)", err)
			return nil, err
		}
	}
	// end acme services

	return service, nil
}
