package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"reflect"
	"time"
)

var (
	errUnsupportedMethod = errors.New("unsupported or disabled challenge method")
	errShutdown          = errors.New("adding challenge record aborted due to shutdown")
	errNameUnavailable   = errors.New("failed to add challenge record due to resource name never becoming free (timeout)")
)

// Provision generates the needed ACME challenge resource (to validate the challenge) and then provisions that
// resource using the Method's provider. It also adds the resource name to the tracking map to prevent duplicate
// resources from attempting to provision at the same time.
func (service *Service) Provision(identifier acme.Identifier, method Method, key acme.AccountKey, token string) (err error) {
	// calculate the needed resource
	resourceName, resourceContent, err := method.validationResource(identifier, key, token)
	if err != nil {
		return err
	}

	// loop to add resource name to those currently provisioned and wait if not available
	// if multiple callers are in the waiting state, it is random which will execute next
	for {
		// add resource name to in use
		alreadyExisted, signal := service.resourceNamesInUse.Add(resourceName)
		// if didn't already exist, break loop and provision
		if !alreadyExisted {
			service.logger.Debugf("added resource name %s to challenge work tracker", resourceName)
			break
		}

		service.logger.Debugf("unable to add resource name %s to challenge work tracker; waiting for name to become free", resourceName)

		// block until name is free, timeout, or shutdown is called
		select {
		// signal channel close indicating resource name should now be available
		case <-signal:
			// continue loop (i.e. retry adding)

		// shutdown - return error
		case <-service.shutdownContext.Done():
			return errShutdown

		// timeout - return error if blocked for an hour (should never happen, but just in case to prevent hang)
		case <-time.After(1 * time.Hour):
			return errNameUnavailable
		}

	}

	// confirm provider is available
	if provider, ok := service.providers[method.Value]; !ok || reflect.ValueOf(provider).IsNil() {
		return errUnsupportedMethod
	}

	// Provision with the appropriate provider
	err = service.providers[method.Value].Provision(resourceName, resourceContent)
	if err != nil {
		return err
	}

	// if using dns-01 method, utilize dnsChecker
	if method.ChallengeType == acme.ChallengeTypeDns01 {
		// check for propagation
		propagated, err := service.dnsChecker.CheckTXTWithRetry(resourceName, resourceContent, 10)
		if err != nil {
			service.logger.Error(err)
			return err
		}

		// if failed to propagate
		if !propagated {
			return dns_checker.ErrDnsRecordNotFound
		}
	}

	return nil
}

// Deprovision generates the needed ACME challenge resource (to validate the challenge) and then deprovisions that
// resource using the Method's provider. It also removes the resource name from the tracking map to indicate the
// name is now available for re-use.
func (service *Service) Deprovision(identifier acme.Identifier, method Method, key acme.AccountKey, token string) (err error) {
	// calculate the needed resource
	resourceName, resourceContent, err := method.validationResource(identifier, key, token)
	if err != nil {
		return err
	}

	// delete resource name from tracker (after the rest of the deprovisioning steps are done or failed)
	defer func() {
		err := service.resourceNamesInUse.Remove(resourceName)
		if err != nil {
			service.logger.Errorf("failed to remove resource name %s from work tracker (%s)", resourceName, err)
		} else {
			service.logger.Debugf("removed resource name %s from challenge work tracker", resourceName)
		}
	}()

	// confirm provider is available
	if provider, ok := service.providers[method.Value]; !ok || reflect.ValueOf(provider).IsNil() {
		return errUnsupportedMethod
	}

	// Deprovision with the appropriate provider
	err = service.providers[method.Value].Deprovision(resourceName, resourceContent)
	if err != nil {
		return err
	}

	return nil
}
