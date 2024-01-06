package challenges

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/randomness"
	"time"

	"github.com/cenkalti/backoff/v4"
)

var (
	errDnsDidntPropagate         = errors.New("solving failed: dns record didn't propagate")
	errChallengeRetriesExhausted = errors.New("solving failed: challenge failed to move to final state")
	errChallengeTypeNotFound     = errors.New("solving failed: provider's challenge type not found in challenges array (possibly trying to use a wildcard with http-01)")
)

// Solve accepts an ACME identifier and a slice of challenges and then solves the challenge using a provider
// for the specific domain. If no provider exists or solving otherwise fails, an error is returned.
func (service *Service) Solve(identifier acme.Identifier, challenges []acme.Challenge, key acme.AccountKey, acmeService *acme.Service) (status string, err error) {
	// get provider for identifier
	provider, err := service.Providers.ProviderFor(identifier)
	if err != nil {
		return "", err
	}

	// range to the correct challenge to solve based on ACME Challenge Type (from provider)
	challengeType := provider.AcmeChallengeType()
	var challenge acme.Challenge
	found := false

	for i := range challenges {
		if challenges[i].Type == challengeType {
			found = true
			challenge = challenges[i]
			break
		}
	}
	if !found {
		return "", errChallengeTypeNotFound
	}

	// vars for provision/deprovision
	domain := identifier.Value
	token := challenge.Token
	keyAuth, err := key.KeyAuthorization(token)
	if err != nil {
		return "", fmt.Errorf("failed to make key auth (%s)", err)
	}

	// provision the needed resource for validation and defer deprovisioning
	// add to wg to ensure deprovision completes during shutdown
	service.shutdownWaitgroup.Add(1)
	err = service.provision(domain, token, keyAuth, provider)
	// do error check after Deprovision to ensure any records that were created
	// get cleaned up, even if Provision errored.

	defer func() {
		err := service.deprovision(domain, token, keyAuth, provider)
		if err != nil {
			service.logger.Errorf("challenge solver deprovision failed (%s)", err)
		}
		// wg done do shutdown can proceed after deprovision
		service.shutdownWaitgroup.Done()
	}()

	// Provision error check
	if err != nil {
		return "", err
	}

	// if using dns-01 provider, utilize dnsChecker
	if challengeType == acme.ChallengeTypeDns01 {
		if service.dnsChecker != nil {
			// get dns record to check
			dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

			// check for propagation
			propagated := service.dnsChecker.CheckTXTWithRetry(dnsRecordName, dnsRecordValue)
			// if failed to propagate
			if !propagated {
				return "", errDnsDidntPropagate
			}
		} else {
			// dnschecker is needed but not configured, shouldn't happen but deal with it just in case
			sleepWait := 240
			service.logger.Error("dns checker is needed by solver but for some reason its not running, manually "+
				"sleeping %s seconds, report this issue to lego dev", sleepWait)
			time.Sleep(time.Duration(sleepWait) * time.Second)
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

	// sleep a little before first check
	time.Sleep(7 * time.Second)

	// monitor challenge status using exponential backoff
	challCheckFunc := func() error {
		// get challenge
		challenge, err = acmeService.GetChallenge(challenge.Url, key)
		if err != nil {
			return err
		}

		// log error if invalid
		if challenge.Status == "invalid" {
			service.logger.Infof("challenge status invalid; acme error: %s", challenge.Error)
		}

		// done if Status has reached a final status
		if challenge.Status == "valid" || challenge.Status == "invalid" {
			return nil
		}

		// not a final status
		return fmt.Errorf("current status (%s) is not a final status", challenge.Status)
	}

	// notify: info log challenge checks
	notifyFunc := func(funcErr error, dur time.Duration) {
		service.logger.Infof("challenge for %s not confirmed finalized (%s), will check again in %s", challenge.Url, funcErr, dur.Round(100*time.Millisecond))
	}

	bo := randomness.BackoffACME(service.shutdownContext)
	err = backoff.RetryNotify(challCheckFunc, bo, notifyFunc)
	// if err returned, retry was exhausted
	if err != nil {
		return "", errChallengeRetriesExhausted
	}

	return challenge.Status, nil
}
