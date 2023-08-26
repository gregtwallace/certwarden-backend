package dns01manual

import (
	"os/exec"
)

// AvailableDomains returns all of the domains that this provider instance can
// provision records for.
func (service *Service) AvailableDomains() []string {
	return service.domains
}

// Provision adds the corresponding DNS record using the script.
func (service *Service) Provision(resourceName, resourceContent string) error {
	// run create script
	// script command
	cmd := service.makeCreateCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get detailed err
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			service.logger.Errorf("dns create script std err: %s", exitErr.Stderr)
		}
		service.logger.Errorf("dns create script error: %s", err)
		return err
	}
	service.logger.Debugf("dns create script output: %s", string(result))

	return nil
}

// Deprovision deletes the corresponding DNS record using the script.
func (service *Service) Deprovision(resourceName, resourceContent string) error {
	// run delete script
	// script command
	cmd := service.makeDeleteCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		// try to get detailed err
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			service.logger.Errorf("dns delete script std err: %s", exitErr.Stderr)
		}
		service.logger.Errorf("dns delete script error: %s", err)
		return err
	}
	service.logger.Debugf("dns delete script output: %s", string(result))

	return nil
}
