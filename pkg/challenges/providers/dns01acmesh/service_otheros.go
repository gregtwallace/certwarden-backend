//go:build !windows

package dns01acmesh

import (
	"certwarden-backend/pkg/datatypes/environment"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var errBashMissing = errors.New("unable to find bash")

// NewService creates a new service
func NewService(app App, cfg *Config) (*Service, error) {
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
