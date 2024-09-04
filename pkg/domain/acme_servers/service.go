package acme_servers

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/httpclient"
	"certwarden-backend/pkg/output"
	"certwarden-backend/pkg/pagination_sort"
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary acme_servers service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetAcmeServerStorage() Storage
	GetHttpClient() *httpclient.Client
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Storage interface for storage functions
type Storage interface {
	GetAllAcmeServers(q pagination_sort.Query) ([]Server, int, error)
	GetOneServerById(acmeServerId int) (Server, error)
	GetOneServerByName(name string) (Server, error)

	PostNewServer(NewPayload) (Server, error)
	PutServerUpdate(UpdatePayload) (Server, error)
	DeleteServer(acmeServerId int) error

	ServerHasAccounts(accountId int) (inUse bool)
}

// Acme service struct
type Service struct {
	logger            *zap.SugaredLogger
	output            *output.Service
	storage           Storage
	httpClient        *httpclient.Client
	shutdownContext   context.Context
	shutdownWaitgroup *sync.WaitGroup
	acmeServers       map[int]*acme.Service // [id]acmeServer
	mu                sync.Mutex
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
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

	// shutdown context
	service.shutdownContext = app.GetShutdownContext()
	if service.shutdownContext == nil {
		return nil, errServiceComponent
	}

	// shutdown waitgroup
	service.shutdownWaitgroup = app.GetShutdownWaitGroup()
	if service.shutdownWaitgroup == nil {
		return nil, errServiceComponent
	}

	// acme services map
	service.acmeServers = make(map[int]*acme.Service)

	// get server list from storage
	servers, _, err := service.storage.GetAllAcmeServers(pagination_sort.Query{})
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
