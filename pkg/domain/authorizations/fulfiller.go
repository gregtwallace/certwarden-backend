package authorizations

import (
	"certwarden-backend/pkg/acme"
	"errors"
	"fmt"
	"sync"
)

// acceptable final (non-error) authorization statuses (see: rfc8555 s 7.1.6)
var finalAuthStatuses = []string{"valid", "invalid", "deactivated", "expired", "revoked"}

// FulfillAuths attempts to validate each of the auth URLs in the slice of auth URLs. It returns an error if any
// auth was not confirmed as in a final state (e.g., 'invalid' auth will not throw an error).
func (service *Service) FulfillAuths(authUrls []string, key acme.AccountKey, acmeService *acme.Service) error {
	// aysnc checking the authz for validity
	var wg sync.WaitGroup
	wgSize := len(authUrls)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// fulfill each auth concurrently
	for i := range authUrls {
		go func(authUrl string) {
			defer wg.Done()
			err := service.fulfillAuth(authUrl, key, acmeService)
			wgErrors <- err
		}(authUrls[i])
	}

	// wait for all auths to do their thing
	wg.Wait()

	// close channel
	close(wgErrors)

	// append any wgErrors to err
	var err error
	for wgErr := range wgErrors {
		if wgErr != nil {
			err = errors.Join(err, wgErr)
		}
	}
	if err != nil {
		return err
	}

	// no errors
	return nil
}

// fulfillAuth attempts to validate an auth URL by calling the challenge solver. If multiple calls are made for
// the same auth, the additional calls will wait in a queue to proceed in turn. An error is returned if the auth
// is not confirmed as in a final state.
func (service *Service) fulfillAuth(authUrl string, key acme.AccountKey, acmeService *acme.Service) error {
	// use a map and signal channels to ensure the same auth is not attempted to be solved simultaneously
	for {
		// add auth
		exists, signal := service.authsWorking.Add(authUrl, make(chan struct{}))

		// if doesn't exist (not working) break from loop and solve
		if !exists {
			break
		}

		// block until the other thread working this auth signals done
		<-signal

		// loop to try and Add to authsWorking again
	}

	// defer removing auth once solving is done
	defer func() {
		// delete func closes the signal channel before returning true
		delFunc := func(key string, signal chan struct{}) bool {
			if key == authUrl {
				close(signal)
				return true
			}
			return false
		}

		deletedOk := service.authsWorking.DeleteFunc(delFunc)
		if !deletedOk {
			service.logger.Errorf("authorizations: failed to remove %s from work tracker", authUrl)
		}
	}()

	// work the auth

	// PaG the authorization
	auth, err := acmeService.GetAuth(authUrl, key)
	if err != nil {
		return err
	}

	// call solver if auth is 'pending' (i.e., needs solving)
	if auth.Status == "pending" {
		err = service.challenges.Solve(auth.Identifier, auth.Challenges, key, acmeService)
		// return error if couldn't solve
		if err != nil {
			return err
		}

		// PaG the authorization again (to confirm state after solve attempt)
		auth, err = acmeService.GetAuth(authUrl, key)
		if err != nil {
			return err
		}
	}

	// check if status is final
	isFinal := false
	for _, finalStatus := range finalAuthStatuses {
		if auth.Status == finalStatus {
			isFinal = true
		}
	}

	// if not final, return error
	if !isFinal {
		return fmt.Errorf("authorizations: auth %s status (%s) is not final", authUrl, auth.Status)
	}

	return nil
}
