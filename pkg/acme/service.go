package acme

import (
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
	// temporarily park this here
	shutdownContext context.Context

	logger       *zap.SugaredLogger
	httpClient   *http.Client
	dirUri       string
	dir          *directory
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
	service.dirUri = dirUri
	service.dir = new(directory)

	// start directory manager
	service.backgroundDirManager(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	// nonce manager
	service.nonceManager = noncemanager.New(service.httpClient, &service.dir.NewNonce)

	service.shutdownContext = app.GetShutdownContext()

	return service, nil
}
