package authorizations

import (
	"errors"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_servers"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary authorizations service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetChallengesService() *challenges.Service
	GetAcmeServerService() *acme_servers.Service
	GetDevMode() bool
}

// service struct
type Service struct {
	logger            *zap.SugaredLogger
	acmeServerService *acme_servers.Service
	challenges        *challenges.Service
	working           *working // tracks auths being worked
	cache             *cache   // tracks results of auths after worked
}

// NewService creates a new service
func NewService(app App) (service *Service, err error) {
	service = new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// acme services
	service.acmeServerService = app.GetAcmeServerService()
	if service.acmeServerService == nil {
		return nil, errServiceComponent
	}

	// challenge solver
	service.challenges = app.GetChallengesService()
	if service.challenges == nil {
		return nil, errServiceComponent
	}

	// initialize working
	service.working = newWorking()

	// initialize cache
	service.cache = newCache()

	return service, nil
}
