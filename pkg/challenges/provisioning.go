package challenges

import (
	"errors"
	"time"
)

var (
	errShutdown        = errors.New("adding challenge record aborted due to shutdown")
	errNameUnavailable = errors.New("failed to add challenge record due to resource name never becoming free (timeout)")
)

// Provision adds the specified ACME Challenge resource name to the in use tracker and then calls the provider
// to provision the actual resource. If the resource name is already in use, it waits until the name is free
// and then proceeds.
func (service *Service) Provision(resourceName, resourceContent string, provider providerService) (err error) {
	// loop to add resourceName to those currently provisioned and wait if not available
	// if multiple callers are in the waiting state, it is random which will execute next
	for {
		// add resourceName to in use
		alreadyExisted, signal := service.resourceNamesInUse.Add(resourceName)
		// if didn't already exist, break loop and provision
		if !alreadyExisted {
			service.logger.Debugf("added resource name %s to challenge work tracker", resourceName)
			break
		}

		service.logger.Debugf("unable to add resource name %s to challenge work tracker; waiting for resource name to become free", resourceName)

		// block until resourceName is free, timeout, or shutdown is called
		select {
		// signal channel close indicating resourceName should now be available
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

	// Provision with the appropriate provider
	err = provider.Provision(resourceName, resourceContent)
	if err != nil {
		return err
	}

	return nil
}

// Deprovision calls the provider to deprovision the actual resource. It then removes the resource name from
// the in use (work) tracker to indicate the name is once again available for use.
func (service *Service) Deprovision(resourceName, resourceContent string, provider providerService) (err error) {
	// delete resource name from tracker (after the rest of the deprovisioning steps are done or failed)
	defer func() {
		err := service.resourceNamesInUse.Remove(resourceName)
		if err != nil {
			service.logger.Errorf("failed to remove resource name %s from work tracker (%s)", resourceName, err)
		} else {
			service.logger.Debugf("removed resource name %s from challenge work tracker", resourceName)
		}
	}()

	// Deprovision with the appropriate provider
	err = provider.Deprovision(resourceName, resourceContent)
	if err != nil {
		return err
	}

	return nil
}
