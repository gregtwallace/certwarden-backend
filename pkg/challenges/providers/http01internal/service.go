package http01internal

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/datatypes/safemap"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary http-01 internal challenge service component is missing")
	errConfigComponent  = errors.New("necessary http-01 config option missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// provider Service struct
type Service struct {
	logger               *zap.SugaredLogger
	shutdownContext      context.Context
	shutdownWaitgroup    *sync.WaitGroup
	stopServerFunc       context.CancelFunc
	stopErrChan          chan error
	port                 int
	provisionedResources *safemap.SafeMap[[]byte]
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is http-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeHttp01
}

// Stop is used for any actions needed prior to deleting this provider. For http-01
// internal, the http server must be shutdown.
func (service *Service) Stop() (err error) {
	// stop server
	service.stopServerFunc()

	// wait for result of server shutdown
	timeoutTimer := time.NewTimer(240 * time.Second)

	select {
	case <-timeoutTimer.C:
		// shutdown timeout
		err = errors.New("http-01 internal server shutdown timed out")
		return err
	case err = <-service.stopErrChan:
		// ensure timer releases resources
		if !timeoutTimer.Stop() {
			<-timeoutTimer.C
		}

		// no-op, proceed to err check
	}

	// common err check (shutdown err = fatal unstable)
	if err != nil {
		err = fmt.Errorf("stop http 01 server failed (%s) leaving http 01 internal provider in an unstable state", err)
		service.logger.Fatal(err)
		// ^ app terminates
		return err
	}

	return nil
}

// Configuration options
type Config struct {
	Port *int `yaml:"port" json:"port"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	if cfg == nil {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// allocate resources map
	service.provisionedResources = safemap.NewSafeMap[[]byte]()

	// set port
	if cfg.Port == nil {
		return nil, errConfigComponent
	}
	service.port = *cfg.Port

	// parent shutdown context
	service.shutdownContext = app.GetShutdownContext()

	// parent shutdown wg
	service.shutdownWaitgroup = app.GetShutdownWaitGroup()

	// start web server for http01 challenges
	err := service.startServer()
	if err != nil {
		return nil, err
	}

	return service, nil
}

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) (err error) {
	// if no config, error
	if cfg == nil {
		return errServiceComponent
	}

	// if port changed, stop server and remake service
	if cfg.Port != nil && *cfg.Port != service.port {
		// stop old server
		err = service.Stop()
		if err != nil {
			return err
		}

		// make new service
		newServ, err := NewService(app, cfg)
		if err != nil {
			// if failed to make, restart old server
			errRestart := service.startServer()
			if errRestart != nil {
				service.logger.Panicf("failed to restart http 01 server leaving http 01 internal provider in an unstable state")
				return errRestart
			}
			return err
		}

		// set content of old pointer so anything with the pointer calls the
		// updated service
		*service = *newServ
	}

	// nothing else to update on service (domains handled by parent pkg)

	return nil
}
