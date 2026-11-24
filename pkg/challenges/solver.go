package challenges

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/randomness"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

var errChallengeRetriesExhausted = errors.New("challenges: solving failed: challenge failed to move to final state (timeout)")

// Solve resolves an authorization to a finalized state by selecting a challenge and using a provider
// to solve that challenge. If no provider exists or solving otherwise fails, an error is returned.
func (service *Service) Solve(authURL string, auth *acme.Authorization, key acme.AccountKey, acmeService *acme.Service) (err error) {
	// decompose auth into pieces (to avoid more extensive refactor for now)
	// TODO: Refactor
	identifier := auth.Identifier
	challenges := auth.Challenges

	// confirm Type is correct (only dns is supported)
	if identifier.Type != acme.IdentifierTypeDns {
		return fmt.Errorf("challenges: acme identifier is type (%s); only 'dns' is supported", string(identifier.Type))
	}

	// identifier value -> provision fqdn
	provisionDomain := service.dnsIDValuetoDomain(identifier.Value)

	// get provider for provision fqdn
	provider, err := service.DNSIdentifierProviders.ProviderFor(provisionDomain)
	if err != nil {
		return err
	}

	challengeType := provider.AcmeChallengeType()
	selectedChall, err := acme.SelectChallenge(challengeType, challenges)
	if err != nil {
		return fmt.Errorf("challenges: error selecting challenge (%w)", err)
	}

	// vars for provision/deprovision
	token := selectedChall.Token
	keyAuth, err := key.KeyAuthorization(token)
	if err != nil {
		return fmt.Errorf("challenges: failed to make key auth (%w)", err)
	}

	// if using an alias, emit debug log message about required CNAME record (or error if challenge
	// type doesn't support a CNAME record)
	cnamePointsFrom, cnamePointsTo, err := acme.DNSChallengeCNAMEInfo(identifier.Value, provisionDomain, challengeType)
	if err != nil {
		return fmt.Errorf("challenges: cname error for identifier '%s', provision domain '%s', and challenge type '%s' (%w)",
			identifier.Value, provisionDomain, challengeType, err)
	}
	if cnamePointsFrom != "" {
		service.logger.Debugf("challenges: alias exists for acme identifier '%s' and manually created cname record pointing from '%s' to '%s' is required",
			identifier.Value, cnamePointsFrom, cnamePointsTo)
	}

	// provision the needed resource for validation and defer deprovisioning
	// add to wg to ensure deprovision completes during shutdown
	service.shutdownWaitgroup.Add(1)
	// Provision with the appropriate provider
	err = service.provision(provisionDomain, token, keyAuth, provider)

	// do error check after Deprovision to ensure any records that were created
	// get cleaned up, even if Provision errored.
	defer func() {
		// don't wait for deprovision to return as it isn't necessary for Solve to
		// be considered concluded
		go func() {
			// wg done do shutdown can proceed after deprovision
			defer service.shutdownWaitgroup.Done()

			err = service.deprovision(provisionDomain, token, keyAuth, provider)
			if err != nil {
				service.logger.Errorf("challenges: deprovision failed (%s)", err)
			}
		}()
	}()

	// Provision error check
	if err != nil {
		return err
	}

	// specified wait time prior to resource check
	wait := provider.PostProvisionResourceWait()
	if wait != time.Duration(0) {
		service.logger.Infof("challenges: waiting to validate %s until %s (delay for propagation of resource)", identifier.Value, time.Now().Add(wait).Format(time.RFC1123))
		select {
		case <-time.After(wait):
			// continue
		case <-service.shutdownContext.Done():
			return errShutdown(provisionDomain)
		}
	} else {
		service.logger.Debugf("challenges: no wait configured for propagation of resource for %s", identifier.Value)
	}

	// Below this point is to inform ACME the challenge is ready to be validated
	// by the server and to subsequently monitor the authorization to be moved to the
	// valid or invalid state.

	// inform ACME that the challenge is ready
	err = acmeService.DoChallengeValidation(selectedChall.Url, key)
	if err != nil {
		return err
	}

	// sleep a little before first check
	time.Sleep(7 * time.Second)

	// monitor auth status using exponential backoff
	// RFC 8555 s. 7.5.1 states to poll the authorization after telling the ACME server
	// the client is ready for challenge validation
	challCheckFunc := func() error {
		// get challenge
		authoriz, err := acmeService.GetAuth(authURL, key)
		if err != nil {
			return err
		}

		// log error if invalid
		if authoriz.Status == "invalid" {
			service.logger.Infof("challenges: authorization %q status invalid", authURL)
		}

		// done if Status has reached a final status
		if authoriz.Status == "valid" || authoriz.Status == "invalid" {
			return nil
		}

		// not a final status
		return fmt.Errorf("challenge %q status (%q) not a final status", authoriz.Status, authURL)
	}

	// notify: info log challenge checks
	notifyFunc := func(funcErr error, dur time.Duration) {
		service.logger.Infof("challenges: %s, will check again in %s", funcErr, dur.Round(100*time.Millisecond))
	}

	bo := randomness.BackoffACME(30*time.Minute, service.shutdownContext)
	err = backoff.RetryNotify(challCheckFunc, bo, notifyFunc)
	// if err returned, retry was exhausted
	if err != nil {
		return errors.Join(errChallengeRetriesExhausted, err)
	}

	return nil
}
