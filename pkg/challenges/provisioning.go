package challenges

import (
	"certwarden-backend/pkg/acme"
	"certwarden-backend/pkg/challenges/providers"
	"certwarden-backend/pkg/randomness"
	"fmt"
	"time"
)

// provision filles the apiMu channel and provisions the resource using the provider. The channel is emptied a few
// seconds after provisioning completes as a way to rate limit these calls.
func (service *Service) provision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// wait for api to be ready for next call (channel available to send empty val to)
	select {
	case service.apiMu <- struct{}{}:
		// proceed
	case <-service.shutdownContext.Done():
		return fmt.Errorf("challenges: provision of %s aborted due to shutdown", domain)
	}

	// defer read from channel (freeing API back up)
	defer func() {
		go func() {
			time.Sleep(time.Duration(2+randomness.GenerateInsecureInt(4)) * time.Second)
			<-service.apiMu
		}()
	}()

	// Provision with the appropriate provider
	err = provider.Provision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	return nil
}

// deprovision fills the apiMu channel and deprovisions the resource using the provider. The channel is emptied one
// second after deprovisioning completes as a way to rate limit these calls.
func (service *Service) deprovision(domain string, token string, keyAuth acme.KeyAuth, provider providers.Service) (err error) {
	// wait for api to be ready for next call (channel available to send empty val to)
	select {
	case service.apiMu <- struct{}{}:
		// proceed
	case <-service.shutdownContext.Done():
		// also proceed - we should deprovision all records
	}

	// defer read from channel (freeing API back up)
	defer func() {
		go func() {
			// keep time low to allow quicker cleanup for shutdown
			time.Sleep(1 * time.Second)
			<-service.apiMu
		}()
	}()

	// Deprovision with the appropriate provider
	err = provider.Deprovision(domain, token, keyAuth)
	if err != nil {
		return err
	}

	return nil
}
