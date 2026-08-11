package providers

import (
	"certwarden-backend/pkg/output"
	"context"
	"errors"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

// application contains functions manager & child providers will need
type application interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetConfigFilenameWithPath() string
	GetShutdownContext() context.Context
	GetHttpClient() *http.Client
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Manager manages the child providers
type Manager struct {
	childApp   application
	logger     *zap.SugaredLogger
	output     *output.Service
	configFile string
	nextId     int
	providers  []*provider
	dP         map[string]*provider // domain -> provider
	mu         sync.RWMutex
}

func MakeManager(app application, cfg Config) (mgr *Manager, err error) {
	// make struct with configs
	mgr = &Manager{
		childApp:   app,
		logger:     app.GetLogger(),
		output:     app.GetOutputter(),
		configFile: app.GetConfigFilenameWithPath(),
		nextId:     0,
		// []*providers
		dP: make(map[string]*provider), // domain -> provider
	}

	// get all provider cfgs as array
	allCfgs := cfg.All()

	// add each provider to manager
	for i := range allCfgs {
		_, err = mgr.unsafeAddProvider(allCfgs[i].internalCfg, allCfgs[i].providerCfg)
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
