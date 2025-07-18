package dns01acmesh

import (
	"certwarden-backend/pkg/acme"
	"errors"
	"os/exec"
)

// Provision adds the requested DNS record.
func (service *Service) Provision(domain string, _ string, keyAuth acme.KeyAuth) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// run create script
	// script command
	cmd := service.makeCreateCommand(dnsRecordName, dnsRecordValue)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			service.logger.Errorf("acme.sh dns provision script std err: %s", exitErr.Stderr)
		}

		service.logger.Errorf("acme.sh dns provision script error: %s", err)
		return err
	}
	service.logger.Debugf("acme.sh dns provision script output: %s", string(result))

	return nil
}

// Deprovision deletes the corresponding DNS record.
func (service *Service) Deprovision(domain string, _ string, keyAuth acme.KeyAuth) error {
	// get dns record
	dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

	// script command
	cmd := service.makeDeleteCommand(dnsRecordName, dnsRecordValue)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			service.logger.Errorf("acme.sh dns deprovision script std err: %s", exitErr.Stderr)
		}

		service.logger.Errorf("acme.sh dns deprovision script error: %s", err)
		return err
	}
	service.logger.Debugf("acme.sh dns deprovision script output: %s", string(result))

	return nil
}
