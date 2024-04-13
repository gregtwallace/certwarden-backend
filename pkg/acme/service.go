package acme

import (
	"certwarden-backend/pkg/acme/nonces"
	"certwarden-backend/pkg/httpclient"
	"context"
	"errors"
	"sync"

	"go.uber.org/zap"
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
	logger       *zap.SugaredLogger
	httpClient   *httpclient.Client
	dirUri       string
	dir          *directory
	nonceManager *nonces.Manager
}

// NewService creates a new service
func NewService(app App, dirUri string) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errors.New("acme: newservice requires valid logger")
	}

	// http client
	service.httpClient = app.GetHttpClient()

	// acme directory
	service.dirUri = dirUri
	service.dir = new(directory)

	// start directory manager
	service.backgroundDirManager(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	// nonce manager
	service.nonceManager = nonces.NewManager(service.httpClient, &service.dir.NewNonce)

	return service, nil
}
