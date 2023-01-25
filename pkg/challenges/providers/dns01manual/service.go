package dns01manual

import (
	"errors"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"legocerthub-backend/pkg/datatypes"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 manual script component is missing")
	errScriptIsDir      = errors.New("dns-01 manual script is a path not a file")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// Accounts service struct
type Service struct {
	logger           *zap.SugaredLogger
	dnsChecker       *dns_checker.Service
	shellPath        string
	createScriptPath string
	deleteScriptPath string
	dnsRecords       *datatypes.SafeMap
}

// Configuration options
type Config struct {
	Enable       *bool  `yaml:"enable"`
	CreateScript string `yaml:"create_script"`
	DeleteScript string `yaml:"delete_script"`
}

// NewService creates a new service
func NewService(app App, config *Config, dnsChecker *dns_checker.Service) (*Service, error) {
	var err error

	// if disabled, return nil and no error
	if !*config.Enable {
		return nil, nil
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// dns checker
	service.dnsChecker = dnsChecker
	if service.dnsChecker == nil {
		return nil, errServiceComponent
	}

	// determine shell (os dependent)
	// powershell
	service.shellPath, err = exec.LookPath("powershell.exe")
	if err != nil {
		// then try bash
		service.shellPath, err = exec.LookPath("bash")
		if err != nil {
			// then try zshell
			service.shellPath, err = exec.LookPath("zsh")
			if err != nil {
				// then try sh
				service.shellPath, err = exec.LookPath("sh")
				if err != nil {
					// failed
					return nil, err
				}
			}
		}
	}

	// verify scripts exist
	// create
	fileInfo, err := os.Stat(config.CreateScript)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, errScriptIsDir
	}
	service.createScriptPath = config.CreateScript

	// delete
	fileInfo, err = os.Stat(config.DeleteScript)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, errScriptIsDir
	}
	service.deleteScriptPath = config.DeleteScript

	// map to hold current dnsRecords
	service.dnsRecords = datatypes.NewSafeMap()

	return service, nil
}
