package acme

import (
	"errors"
	"legocerthub-backend/pkg/acme/nonces"
	"legocerthub-backend/pkg/httpclient"

	"go.uber.org/zap"
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
}

// Acme service struct
type Service struct {
	logger       *zap.SugaredLogger
	httpClient   *httpclient.Client
	dirUri       string
	dir          *acmeDirectory
	nonceManager *nonces.Manager
}

// NewService creates a new service
func NewService(app App, dirUri string) (*Service, error) {
	service := new(Service)
	var err error

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errors.New("acme: newservice requires valid logger")
	}

	// http client
	service.httpClient = app.GetHttpClient()

	// acme directory
	service.dirUri = dirUri
	service.dir = new(acmeDirectory)

	// initial population
	err = service.fetchAcmeDirectory()
	if err != nil {
		return nil, err
	}

	// start go routine to check for periodic updates
	service.backgroundDirManager()

	// nonce manager
	service.nonceManager = nonces.NewManager(service.httpClient, &service.dir.NewNonce)

	return service, nil
}
