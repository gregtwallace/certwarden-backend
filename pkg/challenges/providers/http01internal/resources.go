package http01internal

import "fmt"

// Provision adds a resource to host
func (service *Service) Provision(resourceName, resourceValue string) error {
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
func (service *Service) Deprovision(resourceName, resourceValue string) error {
	// resourceValue is unused but is needed to satisfy an interface in challenges

	// delete entry
	err := service.provisionedResources.DeleteKey(resourceName)
	if err != nil {
		return err
	}

	return nil
}
