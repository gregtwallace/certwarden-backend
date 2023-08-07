package dns01acmesh

import (
	"fmt"
	"os/exec"
)

// Provision adds the resource to the internal tracking map and provisions
// the corresponding DNS record.
func (service *Service) Provision(resourceName string, resourceContent string) error {
	// add to internal map
	exists, existingContent := service.dnsRecords.Add(resourceName, resourceContent)
	// if already exists, but content is different, error
	if exists && existingContent != resourceContent {
		return fmt.Errorf("dns-01 (acme.sh) can't add resource (%s), already exists "+
			"and content does not match", resourceName)
	}

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

// Deprovision removes the resource from the internal tracking map and deletes
// the corresponding DNS record.
func (service *Service) Deprovision(resourceName string, resourceContent string) error {
	// remove from internal map
	err := service.dnsRecords.Delete(resourceName)
	if err != nil {
		service.logger.Errorf("dns-01 (acme.sh) could not remove resource (%s) from "+
			"internal map (%s)", resourceName, err)
		// do not return
	}

	// run delete script
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
