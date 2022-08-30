package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
)

var errWrongIdentifierType = errors.New("acme identifier is not of type dns")

func (service *Service) Provision(identifier acme.Identifier, method Method, key acme.AccountKey, token string) (err error) {
	// verify identifier is the proper type (only dns identifiers are supported)
	if identifier.Type != "dns" {
		return errWrongIdentifierType
	}

	// calculate the resource content based on spec
	resourceName := ""
	resourceContent := ""

	switch method.Type {
	// http-01 provides the keyAuth
	case "http-01":
		// http-01 uses the token as the resource name
		resourceName = token
		// http-01 uses the keyAuth as the resource content
		resourceContent, err = key.KeyAuthorization(token)
		if err != nil {
			return err
		}
	case "dns-01":
		// dns-01 uses "_acme-challenge." prepended to the dns identifier value
		// (e.g. "_acme-challenge.idendifier.example.com") as the resource name
		resourceName = "_acme-challenge." + identifier.Value
		// dns-01 uses the keyAuth's SHA-256 Encoded Hash as the resource content.
		resourceContent, err = key.KeyAuthorizationEndodedSHA256(token)
		if err != nil {
			return err
		}
	default:
		return errUnsupportedMethod
	}

	// Provision with the appropriate provider
	switch method.Value {
	case "http-01-internal":
		service.http01.Provision(resourceName, resourceContent)
	case "dns-01-script":
		// TODO: Support DNS
		service.logger.Errorf("dns-01 unsupported (keyauth hash: %s", resourceContent)
		return errUnsupportedMethod
	default:
		return errUnsupportedMethod
	}

	return nil
}

func (service *Service) Deprovision(identifier acme.Identifier, method Method, token string) (err error) {
	// Deprovision with the appropriate provider
	switch method.Value {
	case "http-01-internal":
		// remove from internal http server
		service.http01.Deprovision(token)
	default:
		return errUnsupportedMethod
	}

	return nil
}
