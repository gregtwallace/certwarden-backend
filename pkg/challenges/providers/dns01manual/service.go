package dns01manual

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
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

// provider Service struct
type Service struct {
	logger           *zap.SugaredLogger
	domains          []string
	shellPath        string
	environmentVars  []string
	createScriptPath string
	deleteScriptPath string
}

// Stop/Start is not needed for this provider. Nothing needs to be stopped or started.
func (service *Service) Stop() error  { return nil }
func (service *Service) Start() error { return nil }

// Configuration options
type Config struct {
	Domains      []string                         `yaml:"domains" json:"domains"`
	Environment  output.RedactedEnvironmentParams `yaml:"environment" json:"environment"`
	CreateScript string                           `yaml:"create_script" json:"create_script"`
	DeleteScript string                           `yaml:"delete_script" json:"delete_script"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config or no domains, error
	if cfg == nil || len(cfg.Domains) <= 0 {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// set supported domains from config
	service.domains = append(service.domains, cfg.Domains...)

	// determine shell (os dependent)
	// powershell
	var err error
	service.shellPath, err = exec.LookPath("powershell.exe")
	if err != nil {
		service.logger.Debugf("unable to find powershell (%s)", err)
		// then try bash
		service.shellPath, err = exec.LookPath("bash")
		if err != nil {
			service.logger.Debugf("unable to find bash (%s)", err)
			// then try zshell
			service.shellPath, err = exec.LookPath("zsh")
			if err != nil {
				service.logger.Debugf("unable to find zshell (%s)", err)
				// then try sh
				service.shellPath, err = exec.LookPath("sh")
				if err != nil {
					service.logger.Debugf("unable to find sh (%s)", err)
					// failed
					return nil, errors.New("unable to find suitable shell")
				}
			}
		}
	}

	// environment vars
	service.environmentVars = cfg.Environment.Unredacted()

	// verify create script exists
	fileInfo, err := os.Stat(cfg.CreateScript)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, errScriptIsDir
	}
	service.createScriptPath = cfg.CreateScript

	// verify delete script exists
	fileInfo, err = os.Stat(cfg.DeleteScript)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, errScriptIsDir
	}
	service.deleteScriptPath = cfg.DeleteScript

	return service, nil
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}
