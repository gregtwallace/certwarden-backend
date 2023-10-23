package updater

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"sync"
	"time"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary updater service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetAppVersion() string
	GetConfigVersion() int
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
	GetOutputter() *output.Service
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// verVersion holds all of the information regarding new version check
// results
type newVersion struct {
	available bool
	info      *versionInfo
	lastCheck time.Time
	mu        sync.RWMutex
}

// Keys service struct
type Service struct {
	logger               *zap.SugaredLogger
	httpClient           *httpclient.Client
	output               *output.Service
	currentVersion       string
	currentConfigVersion int
	checkChannel         Channel
	newVersion           newVersion
}

// Config holds all of the challenge config
type Config struct {
	AutoCheck *bool    `yaml:"auto_check"`
	Channel   *Channel `yaml:"channel"`
}

// NewService creates a new (local LeGo) users service
func NewService(app App, cfg *Config) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// http client
	service.httpClient = app.GetHttpClient()

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// current version
	service.currentVersion = app.GetAppVersion()
	service.currentConfigVersion = app.GetConfigVersion()

	// channel to check
	service.checkChannel = *cfg.Channel

	// start background auto update check service (if enabled)
	if *cfg.AutoCheck {
		service.backgroundChecker(app.GetShutdownContext(), app.GetShutdownWaitGroup())
	}

	return service, nil
}
