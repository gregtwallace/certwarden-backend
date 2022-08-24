package frontend

import (
	"errors"
	"legocerthub-backend/pkg/datatypes"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary frontend service component is missing")
	errImproperStart    = errors.New("attempting to run frontend with config set to disabled")
)

// Config contains specific options for the frontend
type Config struct {
	Enable    *bool `yaml:"enable"`
	HttpsPort *int  `yaml:"https_port"`
	HttpPort  *int  `yaml:"http_port"`
}

// App interface is for connecting to the main app
type App interface {
	GetDevMode() bool
	GetFrontendConfig() Config
	GetHostname() *string
	GetApiPort() *int
	GetLogger() *zap.SugaredLogger
	GetHttpsCert() *datatypes.SafeCert
}

// frontend service struct
type Service struct {
	devMode   bool
	config    Config
	hostname  *string
	apiPort   *int
	logger    *zap.SugaredLogger
	httpsCert *datatypes.SafeCert
}

// NewService creates a new frontend service
func NewService(app App) (*Service, error) {
	service := new(Service)
	var err error

	// devmode
	service.devMode = app.GetDevMode()

	// config
	service.config = app.GetFrontendConfig()
	if !*service.config.Enable {
		return nil, errImproperStart
	}

	// hostname
	service.hostname = app.GetHostname()

	// get port the API (main app) is running on
	service.apiPort = app.GetApiPort()

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// https cert
	service.httpsCert = app.GetHttpsCert()

	// start frontend webserver
	err = service.run()
	if err != nil {
		return nil, err
	}

	return service, nil
}
