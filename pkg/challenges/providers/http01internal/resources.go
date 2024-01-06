package http01internal

import "fmt"

// Provision adds a resource to host
func (service *Service) Provision(_, token, keyAuth string) error {
	// add new entry
	exists, _ := service.provisionedResources.Add(token, []byte(keyAuth))

	// if it already exists, log an error and fail (should never happen if challenges is working
	// properly)
	if exists {
		err := fmt.Errorf("http-01 resource name %s already in use, this should never happen", token)
		service.logger.Error(err)
		return err
	}

	return nil
}

// Deprovision removes a removes a resource from those being hosted
func (service *Service) Deprovision(_, token, _ string) error {

	// delete entry
	err := service.provisionedResources.DeleteKey(token)
	if err != nil {
		return err
	}

	return nil
}
