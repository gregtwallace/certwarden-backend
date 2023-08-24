package http01internal

import "fmt"

// AvailableDomains returns all of the domains that are configured.
// Technically an http01 server can provision a record for any domain, but
// this satisfies what is needed for the main challenges package so it knows
// where to send provisioning commands.
func (service *Service) AvailableDomains() []string {
	return service.domains
}

// Provision adds a resource to host
func (service *Service) Provision(domainName, resourceName, resourceValue string) error {
	// domainName is unused but is needed to satisfy an interface in challenges package

	// add new entry
	exists, _ := service.provisionedResources.Add(resourceName, []byte(resourceValue))

	// if it already exists, log an error and fail (should never happen if challenges is working
	// properly)
	if exists {
		err := fmt.Errorf("http-01 resource name %s already in use, this should never happen", resourceName)
		service.logger.Error(err)
		return err
	}

	return nil
}

// Deprovision removes a removes a resource from those being hosted
func (service *Service) Deprovision(domainName, resourceName, resourceValue string) error {
	// domainName & resourceValue are unused but is needed to satisfy an interface in challenges

	// delete entry
	err := service.provisionedResources.DeleteKey(resourceName)
	if err != nil {
		return err
	}

	return nil
}
