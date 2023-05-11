package updater

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"sync"

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

// Keys service struct
type Service struct {
	logger               *zap.SugaredLogger
	httpClient           *httpclient.Client
	output               *output.Service
	currentVersion       string
	currentConfigVersion int
	checkChannel         Channel
	newVersionAvailable  bool
	newVersionInfo       *versionInfo
	mu                   sync.RWMutex
}

// Config holds all of the challenge config
type Config struct {
	Enable  *bool    `yaml:"enable"`
	Channel *Channel `yaml:"channel"`
}

// NewService creates a new (local LeGo) users service
func NewService(app App, cfg *Config) (*Service, error) {
	// if disabled, return nil and no error
	if !*cfg.Enable {
		return nil, nil
	}

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

	// start background update check service
	service.backgroundChecker(app.GetShutdownContext(), app.GetShutdownWaitGroup())

	return service, nil
}
