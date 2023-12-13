package backup

import (
	"errors"
	"legocerthub-backend/pkg/output"
	"log"
	"path/filepath"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary backup service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetDataStorageRootPath() string
	GetDataStorageLogPath() string
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
}

// Keys service struct
type Service struct {
	cleanDataStorageRootPath   string
	cleanDataStorageLogPath    string
	cleanDataStorageBackupPath string
	logger                     *zap.SugaredLogger
	output                     *output.Service
}

// NewService creates a new service
func NewService(app App) (*Service, error) {
	service := new(Service)

	service.cleanDataStorageRootPath = filepath.Clean(app.GetDataStorageRootPath())
	service.cleanDataStorageLogPath = filepath.Clean(app.GetDataStorageLogPath())
	service.cleanDataStorageBackupPath = filepath.Clean(app.GetDataStorageRootPath() + "/" + dataStorageBackupDirName)

	log.Println(service.cleanDataStorageRootPath)
	log.Println(service.cleanDataStorageLogPath)
	log.Println(service.cleanDataStorageBackupPath)

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
