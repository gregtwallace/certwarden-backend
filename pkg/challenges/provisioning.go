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

	// Provision with the appropriate provider
	switch method {
	case http01Internal:
		err = service.challengeProviders.http01Internal.Provision(resourceName, resourceContent)
	case dns01Script:
		// TODO: Support DNS
		service.logger.Errorf("dns-01 unsupported (keyauth hash: %s", resourceContent)
		return errUnsupportedMethod
	default:
		return errUnsupportedMethod
	}

	// centralize error check from provisioning
	if err != nil {
		return err
	}

	return nil
}

// Deprovision removes the ACME challenge resource from the Method's
// provider.
func (service *Service) Deprovision(identifier acme.Identifier, method Method, token string) (err error) {
	// Deprovision with the appropriate provider
	switch method {
	case http01Internal:
		// remove from internal http server
		err = service.challengeProviders.http01Internal.Deprovision(token)
	default:
		return errUnsupportedMethod
	}

	if err != nil {
		return err
	}

	return nil
}
