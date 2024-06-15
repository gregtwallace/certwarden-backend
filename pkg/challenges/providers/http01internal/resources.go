package http01internal

import (
	"certwarden-backend/pkg/acme"
	"fmt"
)

// Provision adds a resource to host
func (service *Service) Provision(_ string, token string, keyAuth acme.KeyAuth) error {
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
func (service *Service) Deprovision(_ string, token string, _ acme.KeyAuth) error {
	// delete entry
	delFunc := func(tokenKey string, _ []byte) bool {
		return tokenKey == token
	}

	deleteOk := service.provisionedResources.DeleteFunc(delFunc)
	if !deleteOk {
		return fmt.Errorf("http-01 resource %s failed to delete", token)
	}

	return nil
}
