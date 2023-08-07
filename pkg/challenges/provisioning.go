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
	errUnavailableName   = errors.New("failed to add challenge record due to name availability time out")
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

	// exists tracker for loop (start as true which makes loop execute)
	exists := true

	// retry loop
	for i := 1; exists; i++ {
		// lock, read, write if name is available, unlock
		service.resourceNamesProvisioned.mu.Lock()
		_, exists = service.resourceNamesProvisioned.names[resourceName]
		if !exists {
			service.resourceNamesProvisioned.names[resourceName] = struct{}{}
		}
		service.resourceNamesProvisioned.mu.Unlock()

		// if writing succeeded, done with loop
		if !exists {
			break
		}

		// if loop failed 12th time, error out
		if i >= 12 {
			service.logger.Error(errUnavailableName)
			return errUnavailableName
		}

		// sleep or cancel/error if shutdown is called
		select {
		case <-service.shutdownContext.Done():
			// cancel/error if shutting down
			return errShutdown

		case <-time.After(time.Duration(i) * 15 * time.Second):
			// sleep and retry
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

// Deprovision removes the ACME challenge resource from the Method's provider.
func (service *Service) Deprovision(identifier acme.Identifier, method Method, key acme.AccountKey, token string) (err error) {
	// calculate the needed resource
	resourceName, resourceContent, err := method.validationResource(identifier, key, token)
	if err != nil {
		return err
	}

	// delete resource from tracker (after the rest of the func is done)
	defer func() {
		// lock, delete, unlock
		service.resourceNamesProvisioned.mu.Lock()
		delete(service.resourceNamesProvisioned.names, resourceName)
		service.resourceNamesProvisioned.mu.Unlock()
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
