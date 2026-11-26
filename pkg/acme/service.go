package acme

import (
	"certwarden-backend/pkg/acme/internal/directory"
	"certwarden-backend/pkg/acme/internal/noncemanager"
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
	// TODO: remove this
	shutdownContext context.Context

	logger       *zap.SugaredLogger
	httpClient   *http.Client
	dir          *directory.DirectoryMgr
	nonceManager *noncemanager.Manager
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
	service.dir = directory.New(app.GetLogger(), app.GetHttpClient(), dirUri)

	// nonce manager
	service.nonceManager = noncemanager.New(service.httpClient, service.dir)

	service.shutdownContext = app.GetShutdownContext()

	return service, nil
}
