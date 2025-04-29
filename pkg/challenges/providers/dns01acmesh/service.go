package dns01acmesh

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/datatypes/environment"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 acme.sh component is missing")
	errBashMissing      = errors.New("unable to find bash")
	errWindows          = errors.New("acme.sh is not supported in windows, disable it")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
}

// provider Service struct
type Service struct {
	logger            *zap.SugaredLogger
	shellPath         string
	acmeShPath        string
	dnsHook           string
	environmentParams *environment.Params
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// Configuration options
type Config struct {
	AcmeShPath  string   `yaml:"acme_sh_path" json:"acme_sh_path"`
	Environment []string `yaml:"environment" json:"environment"`
	DnsHook     string   `yaml:"dns_hook" json:"dns_hook"`
}

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
	// error and fail if trying to run on windows
	if runtime.GOOS == "windows" {
		return nil, errWindows
	}

	// if no config, error
	if cfg == nil {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// bash is required (due to using `source`)
	var err error
	service.shellPath, err = exec.LookPath("bash")
	if err != nil {
		return nil, errBashMissing
	}

	// hook name (needed for funcs) & path
	service.dnsHook = cfg.DnsHook
	service.acmeShPath = cfg.AcmeShPath

	// check for the needed dns script in custom folder
	_, err = os.Stat(service.acmeShPath + dnsApiCwPath + "/" + service.dnsHook + ".sh")
	if err != nil {
		return nil, fmt.Errorf("acme.sh: erorr opening dns script (%s)", err)
	}

	// environment vars
	var invalidParams []string
	service.environmentParams, invalidParams = environment.NewParams(cfg.Environment)
	if len(invalidParams) > 0 {
		service.logger.Errorf("dns-01 acme.sh some environment param(s) invalid and won't be used (%s)", invalidParams)
	}

	return service, nil
}

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) error {
	// if no config, error
	if cfg == nil {
		return errServiceComponent
	}

	// don't need to do anything with "old" Service, just set a new one
	newServ, err := NewService(app, cfg)
	if err != nil {
		return err
	}

	// set content of old pointer so anything with the pointer calls the
	// updated service
	*service = *newServ

	return nil
}
