package backup

import (
	"context"
	"errors"
	"fmt"
	"legocerthub-backend/pkg/output"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary backup service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDataStorageRootPath() string
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	LockSQLForBackup() (unlockFunc func(), err error)
	GetShutdownContext() context.Context
	GetShutdownWaitGroup() *sync.WaitGroup
}

// Keys service struct
type Service struct {
	cleanDataStorageRootPath   string
	cleanDataStorageBackupPath string
	lockSQLForBackup           func() (unlockFunc func(), err error)
	logger                     *zap.SugaredLogger
	output                     *output.Service
	config                     *Config
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	service.cleanDataStorageRootPath = filepath.Clean(app.GetDataStorageRootPath())
	service.cleanDataStorageBackupPath = filepath.Clean(app.GetDataStorageRootPath() + "/" + dataStorageBackupDirName)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// create backup storage folder, if doesn't exist
	err := os.MkdirAll(service.cleanDataStorageBackupPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("backup: failed to make directory for on disk backups (%s)", err)
	}

	// storage lock func
	service.lockSQLForBackup = app.LockSQLForBackup

	// do not start auto service
	// must be started later in app (after config is read)
	service.config = &Config{}
	service.config.Enabled = new(bool)
	*service.config.Enabled = false

	return service, nil
}
