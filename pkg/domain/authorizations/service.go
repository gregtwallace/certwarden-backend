package authorizations

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/challenges/providers/http01internal"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary authorizations service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetDevMode() bool
	GetHttp01InternalConfig() *http01internal.Config
}

// service struct
type Service struct {
	logger      *zap.SugaredLogger
	acmeProd    *acme.Service
	acmeStaging *acme.Service
	challenges  *challenges.Service
	working     *working // tracks auths being worked
	cache       *cache   // tracks results of auths after worked
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
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// challenge solver
	service.challenges, err = challenges.NewService(app)
	if err != nil || service.challenges == nil {
		return nil, errServiceComponent
	}

	// initialize working
	service.working = newWorking()

	// initialize cache
	service.cache = newCache()

	return service, nil
}
