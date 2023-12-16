package backup

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"path/filepath"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary backup service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDataStorageRootPath() string
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
}

// Keys service struct
type Service struct {
	cleanDataStorageRootPath   string
	cleanDataStorageBackupPath string
	logger                     *zap.SugaredLogger
	output                     *output.Service
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

	return service, nil
}
