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

	case dns01Cloudflare:
		err = service.challengeProviders.dns01Cloudflare.Provision(resourceName, resourceContent)

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
	// calculate the needed resource name
	resourceName, err := method.validationResourceName(identifier, token)
	if err != nil {
		return err
	}

	// Deprovision with the appropriate provider
	switch method {
	case http01Internal:
		err = service.challengeProviders.http01Internal.Deprovision(resourceName)

	case dns01Cloudflare:
		err = service.challengeProviders.dns01Cloudflare.Deprovision(resourceName)

	default:
		return errUnsupportedMethod
	}

	// central error check
	if err != nil {
		return err
	}

	return nil
}
