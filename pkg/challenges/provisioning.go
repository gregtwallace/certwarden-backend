package challenges

import (
	"legocerthub-backend/pkg/acme"
)

// Provision generates the needed ACME challenge resource (to validate
// the challenge) and then provisions that resource using the Method's
// provider.
func (service *Service) Provision(identifier acme.Identifier, method Method, key acme.AccountKey, token string) (err error) {
	// calculate the needed resource
	resourceName, resourceContent, err := method.validationResource(identifier, key, token)
	if err != nil {
		return err
	}

	// confirm provider is available
	if service.providers[method] == nil {
		return errUnsupportedMethod
	}

	// Provision with the appropriate provider
	err = service.providers[method].Provision(resourceName, resourceContent)
	if err != nil {
		return err
	}

	return nil
}

// Deprovision removes the ACME challenge resource from the Method's
// provider.
func (service *Service) Deprovision(identifier acme.Identifier, method Method, token string) (err error) {
	// calculate the needed resource name
	resourceName, err := method.validationResourceName(identifier, token)
	if err != nil {
		return err
	}

	// confirm provider is available
	if service.providers[method] == nil {
		return errUnsupportedMethod
	}

	// Deprovision with the appropriate provider
	err = service.providers[method].Deprovision(resourceName)
	if err != nil {
		return err
	}

	return nil
}
