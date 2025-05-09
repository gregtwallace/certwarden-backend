package challenges

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/challenges/providers"
	"context"
	"errors"
	"fmt"
)

func errShutdown(domain string) error {
	return fmt.Errorf("challenges: %s aborting due to shutdown", domain)
}

// provision filles the apiMu channel and provisions the resource using the provider. The channel is emptied a few
// seconds after provisioning completes as a way to rate limit these calls.
func (service *Service) provision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// impose rate limit
	err = service.apiRateLimiter.Wait(service.shutdownContext)
	if err != nil {
		// if shutdown, return that err
		if errors.Is(err, context.Canceled) {
			return errShutdown(domain)
		}
		// otherwise return error as-is (this shouldn't happen though)
		service.logger.Errorf("challenges: unexpected context error (%v) for domain %s", err, domain)
		return err
	}

	// Provision with the appropriate provider
	err = provider.Provision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	service.logger.Infof("challenges: provisioned domain %s", domain)
	service.logger.Debugf("challenges: domain %s used token %s (key auth: %s)", domain, token, keyAuth)
	return nil
}

// deprovision fills the apiMu channel and deprovisions the resource using the provider. The channel is emptied one
// second after deprovisioning completes as a way to rate limit these calls.
func (service *Service) deprovision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// impose rate limit, but use background context
	// background context ensures we try to deprovision all records
	err = service.apiRateLimiter.Wait(context.Background())
	if err != nil {
		// return error as-is (this should, in theory, never error)
		service.logger.Errorf("challenges: unexpected context error (%v) for domain %s", err, domain)
		return err
	}

	// Deprovision with the appropriate provider
	err = provider.Deprovision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	service.logger.Infof("challenges: deprovisioned domain %s", domain)
	service.logger.Debugf("challenges: domain %s previously used token %s (key auth: %s)", domain, token, keyAuth)
	return nil
}
