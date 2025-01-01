package acme_servers

import (
	"context"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// functions so that acme_servers.Service satisfies the App interface
// contained within acme pkg. This allows acme_servers to start up
// new acme.Service
func (serv *Service) GetLogger() *zap.SugaredLogger {
	return serv.logger
}

func (serv *Service) GetHttpClient() *http.Client {
	return serv.httpClient
}

func (serv *Service) GetShutdownContext() context.Context {
	return serv.shutdownContext
}

func (serv *Service) GetShutdownWaitGroup() *sync.WaitGroup {
	return serv.shutdownWaitgroup
}
