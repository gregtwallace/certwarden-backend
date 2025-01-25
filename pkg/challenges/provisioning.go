package challenges

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/challenges/providers"
	"errors"
	"time"
)

var (
	errShutdown        = errors.New("challenges: solving aborted due to challenges shutdown")
	errNameUnavailable = errors.New("challenges: failed to add challenge record due to resource name never becoming free (timeout)")
)

// Provision adds the specified domain name to the in use tracker and then calls the provider
// to provision the actual resource. If the domain name is already in use, it waits until the name is free
// and then proceeds.
func (service *Service) provision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// loop to add domain to those currently provisioned and wait if not available
	// if multiple callers are in the waiting state, it is random which will execute next
	for {
		// add domain to in use
		alreadyExists, signal := service.resourcesInUse.Add(domain, make(chan struct{}))
		// if didn't already exist, break loop and provision
		if !alreadyExists {
			service.logger.Debugf("challenges: added resource for %s to work tracker", domain)
			break
		}

		service.logger.Debugf("challenges: unable to add resource for %s to work tracker; waiting for resource name to become free", domain)

		// block until domain is free, timeout, or shutdown is called
		timeoutTimer := time.NewTimer(1 * time.Hour)

		select {
		// signal channel close indicating domain should now be available
		case <-signal:
			// ensure timer releases resources
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}

			// continue loop (i.e. retry adding)

		// shutdown - return error
		case <-service.shutdownContext.Done():
			// ensure timer releases resources
			if !timeoutTimer.Stop() {
				<-timeoutTimer.C
			}

			return errShutdown

		// timeout - return error if blocked too long (should never happen, but just in case to prevent hang)
		case <-timeoutTimer.C:
			return errNameUnavailable
		}
	}

	// Provision with the appropriate provider
	err = provider.Provision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	return nil
}

// Deprovision calls the provider to deprovision the actual resource. It then removes the resource name from
// the in use (work) tracker to indicate the name is once again available for use.
func (service *Service) deprovision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// delete resource name from tracker (after the rest of the deprovisioning steps are done or failed)
	// sleep prior to actually closing the signal channel. in the event another solve is waiting for the
	// same resource name, we wait to insert a delay prior to the next solve starting. without this delay
	// the ACME server may mistakenly still see the previous solve value and error. this didn't seem to
	// be an issue with Let's Encrypt and Wild Cards, but once CNAME aliases were added the problem was
	// consistently repeatable thus making this delay necessary
	defer func() {
		sleepSecs := 60
		delFunc := func(key string, signal chan struct{}) bool {
			if key == domain {
				service.logger.Debugf("challenges: sleeping %d seconds before releasing resource %s from work tracker", sleepSecs, domain)
				timeoutTimer := time.NewTimer(time.Duration(sleepSecs) * time.Second)
				select {
				// shutdown - skip waiting
				case <-service.shutdownContext.Done():
					if !timeoutTimer.Stop() {
						<-timeoutTimer.C
					}
				// wait after deprovision before closing signal channel
				case <-timeoutTimer.C:
				}

				// close and return true to indicate do deletion
				close(signal)
				return true
			}
			return false
		}

		deletedOk := service.resourcesInUse.DeleteFunc(delFunc)
		if !deletedOk {
			service.logger.Errorf("challenges: failed to remove resource for %s from work tracker (%s)", domain, err)
		} else {
			service.logger.Debugf("challenges: removed resource for %s from work tracker", domain)
		}
	}()

	// Deprovision with the appropriate provider
	err = provider.Deprovision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	return nil
}
