package dns01acmesh

import (
	"os/exec"
)

// Provision adds the requested DNS record.
func (service *Service) Provision(resourceName, resourceContent string) error {
	// domain is not used

	// run create script
	// script command
	cmd := service.makeCreateCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get detailed err
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			service.logger.Errorf("acme.sh dns create script std err: %s", exitErr.Stderr)
		}
		service.logger.Errorf("acme.sh dns create script error: %s", err)
		return err
	}
	service.logger.Debugf("acme.sh dns create script output: %s", string(result))

	return nil
}

// Deprovision deletes the corresponding DNS record.
func (service *Service) Deprovision(resourceName, resourceContent string) error {
	// script command
	cmd := service.makeDeleteCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get detailed err
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			service.logger.Errorf("acme.sh dns delete script std err: %s", exitErr.Stderr)
		}
		service.logger.Errorf("acme.sh dns delete script error: %s", err)
		return err
	}
	service.logger.Debugf("acme.sh dns delete script output: %s", string(result))

	return nil
}
