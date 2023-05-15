package dns01manual

import (
	"fmt"
)

// Provision adds the resource to the internal tracking map and provisions
// the corresponding DNS record on Cloudflare.
func (service *Service) Provision(resourceName string, resourceContent string) error {
	// add to internal map
	exists, existingContent := service.dnsRecords.Add(resourceName, resourceContent)
	// if already exists, but content is different, error
	if exists && existingContent != resourceContent {
		return fmt.Errorf("dns-01 (manual script) can't add resource (%s), already exists "+
			"and content does not match", resourceName)
	}

	// run create script
	// script command
	cmd := service.makeCreateCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		service.logger.Errorf("dns create script error: %s", err)
		return err
	}
	service.logger.Debugf("dns create script output: %s", string(result))

	return nil
}

// Deprovision removes the resource from the internal tracking map and deletes
// the corresponding DNS record on Cloudflare.
func (service *Service) Deprovision(resourceName string, resourceContent string) error {
	// remove from internal map
	err := service.dnsRecords.Delete(resourceName)
	if err != nil {
		service.logger.Errorf("dns-01 (manual script) could not remove resource (%s) from "+
			"internal map", resourceName)
		// do not return
	}

	// run delete script
	// script command
	cmd := service.makeDeleteCommand(resourceName, resourceContent)

	// run script command
	result, err := cmd.Output()
	if err != nil {
		service.logger.Errorf("dns delete script error: %s", err)
		return err
	}
	service.logger.Debugf("dns delete script output: %s", string(result))

	return nil
}
