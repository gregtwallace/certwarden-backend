package providers

import (
	"context"
	"errors"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"
	"sync"

	"go.uber.org/zap"
)

// application contains functions manager & child providers will need
type application interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetConfigFilename() string
	GetShutdownContext() context.Context
	GetHttpClient() *httpclient.Client
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Manager manages the child providers
type Manager struct {
	childApp   application
	logger     *zap.SugaredLogger
	output     *output.Service
	configFile string
	nextId     int
	dP         map[string]*provider   // domain -> provider
	pD         map[*provider][]string // provider -> []domain
	mu         sync.RWMutex
}

func MakeManager(app application, cfg Config) (mgr *Manager, err error) {
	// make struct with configs
	mgr = &Manager{
		childApp:   app,
		logger:     app.GetLogger(),
		output:     app.GetOutputter(),
		configFile: app.GetConfigFilename(),
		nextId:     0,
		dP:         make(map[string]*provider),   // domain -> provider
		pD:         make(map[*provider][]string), // provider -> []domain
	}

	// get all provider cfgs as array
	allCfgs := cfg.All()

	// add each provider to manager
	for i := range allCfgs {
		_, err = mgr.unsafeAddProvider(allCfgs[i])
		if err != nil {
			return nil, err
		}
	}

	// verify at least one domain / provider exists
	if len(mgr.dP) <= 0 {
		return nil, errors.New("no challenge providers are properly configured (at least one must be enabled)")
	}

	return mgr, nil
}
