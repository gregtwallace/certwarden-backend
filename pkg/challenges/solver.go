package challenges

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/randomness"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// minimum time to wait for challenge solution to propagate after provisioning is completed
const mandatoryWaitAfterProvision = 3 * time.Minute

var (
	errDnsDidntPropagate         = errors.New("challenges: solving failed: dns record didn't propagate")
	errChallengeRetriesExhausted = errors.New("challenges: solving failed: challenge failed to move to final state (timeout)")
	errChallengeTypeNotFound     = errors.New("challenges: solving failed: provider's challenge type not found in challenges array (possibly trying to use a wildcard with http-01)")
)

// Solve accepts an ACME identifier and a slice of challenges and then solves the challenge using a provider
// for the specific domain. If no provider exists or solving otherwise fails, an error is returned.
func (service *Service) Solve(identifier acme.Identifier, challenges []acme.Challenge, key acme.AccountKey, acmeService *acme.Service) (err error) {
	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return fmt.Errorf("challenges: acme identifier is type (%s); only 'dns' is supported", string(identifier.Type))
	}

	// identifier value -> fqdn
	domain := service.dnsIDValuetoDomain(identifier.Value)
	if domain != identifier.Value {
		service.logger.Debugf("challenges: alias exists for acme identifier `%s` and will provision to `%s`", identifier.Value, domain)
	}

	// get provider for fqdn
	provider, err := service.DNSIdentifierProviders.ProviderFor(domain)
	if err != nil {
		return err
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
		return errChallengeTypeNotFound
	}

	// vars for provision/deprovision
	token := challenge.Token
	keyAuth, err := key.KeyAuthorization(token)
	if err != nil {
		return fmt.Errorf("challenges: failed to make key auth (%s)", err)
	}

	// if using an alias, ensure the proper CNAME record exists
	if domain != identifier.Value {
		// exact cname domain depends on challenge type
		cnamePointsFrom := ""
		cnamePointsTo := ""
		if challengeType == acme.ChallengeTypeDns01 {
			cnamePointsFrom = "_acme-challenge." + identifier.Value
			cnamePointsTo = "_acme-challenge." + domain
		} else if challengeType == acme.ChallengeTypeHttp01 {
			cnamePointsFrom = identifier.Value
			cnamePointsTo = domain
		} else {
			return fmt.Errorf("challenges: challenge type %s doesnt support using a domain alias (domain: %s)", challengeType, domain)
		}

		exists := service.dnsChecker.CheckCNAME(cnamePointsFrom, cnamePointsTo)
		if !exists {
			return fmt.Errorf("challenges: cname record %s doesn't exist or doesn't point to %s", cnamePointsFrom, cnamePointsTo)
		}

		service.logger.Debugf(("challenges: cname record %s found and points to %s"), cnamePointsFrom, cnamePointsTo)
	}

	// provision the needed resource for validation and defer deprovisioning
	// add to wg to ensure deprovision completes during shutdown
	service.shutdownWaitgroup.Add(1)
	// Provision with the appropriate provider
	err = service.provision(domain, token, keyAuth, provider)

	// do error check after Deprovision to ensure any records that were created
	// get cleaned up, even if Provision errored.
	defer func() {
		// don't wait for deprovision to return as it isn't necessary for Solve to
		// be considered concluded
		go func() {
			// wg done do shutdown can proceed after deprovision
			defer service.shutdownWaitgroup.Done()

			err = service.deprovision(domain, token, keyAuth, provider)
			if err != nil {
				service.logger.Errorf("challenges: deprovision failed (%s)", err)
			}
		}()
	}()

	// Provision error check
	if err != nil {
		return err
	}
	// measure time after provisioning is completed
	provisionDoneTime := time.Now()

	// if using dns-01 provider, utilize dnsChecker
	if challengeType == acme.ChallengeTypeDns01 {
		// get dns record to check
		dnsRecordName, dnsRecordValue := acme.ValidationResourceDns01(domain, keyAuth)

		// check for propagation
		propagated := service.dnsChecker.CheckTXTWithRetry(dnsRecordName, dnsRecordValue)
		// if failed to propagate
		if !propagated {
			return errDnsDidntPropagate
		}
	}

	// ensure mandatory minimum wait after provision completed
	waitUntilTime := provisionDoneTime.Add(mandatoryWaitAfterProvision)
	if time.Now().Before(waitUntilTime) {
		service.logger.Debugf("challenges: minimum provisioning delay not met, waiting to check %s until %s", identifier.Value, waitUntilTime.Format(time.RFC1123))
		select {
		case <-time.After(time.Until(waitUntilTime)):
			// continue
		case <-service.shutdownContext.Done():
			return errShutdown(domain)
		}
	}

	// Below this point is to inform ACME the challenge is ready to be validated
	// by the server and to subsequently monitor the challenge to be moved to the
	// valid or invalid state.

	// inform ACME that the challenge is ready
	challenge, err = acmeService.InstructServerToValidateChallenge(challenge.Url, key)
	if err != nil {
		return err
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
			service.logger.Infof("challenges: challenge %s status invalid; acme error: %s", challenge.Url, challenge.Error)
		}

		// done if Status has reached a final status
		if challenge.Status == "valid" || challenge.Status == "invalid" {
			return nil
		}

		// not a final status
		return fmt.Errorf("challenge %s status (%s) not a final status", challenge.Status, challenge.Url)
	}

	// notify: info log challenge checks
	notifyFunc := func(funcErr error, dur time.Duration) {
		service.logger.Infof("challenges: %s, will check again in %s", funcErr, dur.Round(100*time.Millisecond))
	}

	bo := randomness.BackoffACME(service.shutdownContext)
	err = backoff.RetryNotify(challCheckFunc, bo, notifyFunc)
	// if err returned, retry was exhausted
	if err != nil {
		return errors.Join(errChallengeRetriesExhausted, err)
	}

	return nil
}
