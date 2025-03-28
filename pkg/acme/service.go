package acme

import (
	"certwarden-backend/pkg/acme/nonces"
	"context"
	"errors"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *http.Client
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Acme service struct
type Service struct {
	logger       *zap.SugaredLogger
	httpClient   *http.Client
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
	service.nonceManager = nonces.NewManager(service.httpClient, app.GetShutdownContext(), &service.dir.NewNonce)

	return service, nil
}
