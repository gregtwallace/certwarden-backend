package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges/dns_checker"
	"time"
)

var (
	errChallengeRetriesExhausted = errors.New("challenge failed (out of retries)")
	errChallengeTypeNotFound     = errors.New("provider's challenge type not found in challenges array (possibly trying to use a wildcard with http-01)")
)

// Solve accepts an ACME identifier and a slice of challenges and then solves the challenge using a provider
// for the specific domain. If no provider exists or solving otherwise fails, an error is returned.
func (service *Service) Solve(identifier acme.Identifier, challenges []acme.Challenge, key acme.AccountKey, acmeService *acme.Service) (status string, err error) {
	// get provider for identifier
	provider, err := service.providers.provider(identifier)
	if err != nil {
		return "", err
	}

	// range to the correct challenge to solve based on ACME Challenge Type (from provider)
	providerChallengeType := provider.AcmeChallengeType()
	var challenge acme.Challenge
	found := false

	for i := range challenges {
		if challenges[i].Type == providerChallengeType {
			found = true
			challenge = challenges[i]
		}
	}
	if !found {
		return "", errChallengeTypeNotFound
	}

	// create resource name and value
	resourceName, resourceContent, err := challenge.ValidationResource(identifier, key)
	if err != nil {
		return "", err
	}

	// provision the needed resource for validation and defer deprovisioning
	err = service.Provision(resourceName, resourceContent, provider)
	// do error check after Deprovision to ensure any records that were created
	// get cleaned up, even if Provision errored.

	defer func() {
		err := service.Deprovision(resourceName, resourceContent, provider)
		if err != nil {
			service.logger.Errorf("challenge solver deprovision failed (%s)", err)
		}
	}()

	// Provision error check
	if err != nil {
		return "", err
	}

	// if using dns-01 provider, utilize dnsChecker
	if providerChallengeType == acme.ChallengeTypeDns01 {
		// check for propagation
		propagated, err := service.dnsChecker.CheckTXTWithRetry(resourceName, resourceContent, 10)
		if err != nil {
			service.logger.Error(err)
			return "", err
		}

		// if failed to propagate
		if !propagated {
			return "", dns_checker.ErrDnsRecordNotFound
		}
	}

	// Below this point is to inform ACME the challenge is ready to be validated
	// by the server and to subsequently monitor the challenge to be moved to the
	// valid or invalid state.

	// inform ACME that the challenge is ready
	_, err = acmeService.ValidateChallenge(challenge.Url, key)
	if err != nil {
		return "", err
	}

	// monitor for processing to complete (max 5 tries, 20 seconds apart each)
	for i := 1; i <= 5; i++ {
		// sleep to allow ACME time to process
		// cancel/error if shutdown is called
		select {
		case <-service.shutdownContext.Done():
			// cancel/error if shutting down
			return "", errors.New("cloudflare dns provisioning canceled due to shutdown")

		case <-time.After(20 * time.Second):
			// sleep and retry
		}

		// get challenge and check for error or final Statuses
		challenge, err = acmeService.GetChallenge(challenge.Url, key)
		if err != nil {
			return "", err
		}

		// return Status if it has reached a final status
		if challenge.Status == "valid" {
			return challenge.Status, nil
		} else if challenge.Status == "invalid" {
			service.logger.Debug(challenge.Error)
			return challenge.Status, nil
		}
		// else repeat loop
	}

	// loop ended without reaching valid or invalid Status
	return "", errChallengeRetriesExhausted
}
