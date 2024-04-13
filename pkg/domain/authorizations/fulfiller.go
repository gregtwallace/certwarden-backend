package authorizations

import (
	"certwarden-backend/pkg/acme"
	"errors"
	"sync"
)

var errAuthPending = errors.New("one or more auths are still in 'pending' status")

// FulfillAuths attempts to validate each of the auth URLs in the slice of auth URLs. It returns 'valid' Status if all auths were
// determined to be 'valid'. It returns 'invalid' if any of the auths were determined to be in any state other than valid or pending.
// It returns an error if any of the auth Statuses could not be determined or if any are still in pending.
func (service *Service) FulfillAuths(authUrls []string, key acme.AccountKey, acmeService *acme.Service) (status string, err error) {
	// aysnc checking the authz for validity
	var wg sync.WaitGroup
	wgSize := len(authUrls)

	wg.Add(wgSize)
	wgStatuses := make(chan string, wgSize)
	wgErrors := make(chan error, wgSize)

	// fulfill each auth concurrently
	// TODO: Add context to cancel everything if any auth fails / invalid?
	for i := range authUrls {
		go func(authUrl string) {
			defer wg.Done()
			status, err = service.fulfillAuth(authUrl, key, acmeService)
			wgStatuses <- status
			wgErrors <- err
		}(authUrls[i])
	}

	// wait for all auths to do their thing
	wg.Wait()

	// close channel
	close(wgStatuses)
	close(wgErrors)

	// check for errors, returns first encountered error only
	for err = range wgErrors {
		if err != nil {
			return "", err
		}
	}

	// check if all auths are valid, if not return invalid
	for status = range wgStatuses {
		if status == "pending" {
			return "", errAuthPending
		} else if status != "valid" {
			return "invalid", nil
		}
	}

	// if looped through all statuses and confirmed all 'valid'
	return "valid", nil
}

// fulfillAuth attempts to validate an auth URL using the specified method. It will either respond from cache
// or call an authWorker.  An error is returned if the auth status could not be determined.
func (service *Service) fulfillAuth(authUrl string, key acme.AccountKey, acmeService *acme.Service) (status string, err error) {
	// add authUrl to working and call a worker, if the authUrl is already being worked,
	// block and return the cached result. If the cached result is an error, try to work
	// the auth again.
	for {
		exists, signal := service.authsWorking.Add(authUrl, make(chan struct{}))
		// if doesn't exist (not working) break from loop and call worker
		if !exists {
			break
		}

		// block until this auth's work (on other thread) is complete
		<-signal

		// read results of the other thread from the cache
		status, err = service.cache.read(authUrl)

		// if no error, return the status from cache
		if err == nil {
			return status, nil
		}

		// if there was an error in cache, loop repeats and authUrl is added to working again
	}

	// defer removing auth once it has been worked
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
			service.logger.Error(err)
		}
	}()

	// work the auth
	status, err = service.authWorker(authUrl, key, acmeService)

	// cache result &
	// error check
	if err != nil {
		service.cache.add(authUrl, "", err)
		return "", err
	}

	service.cache.add(authUrl, status, nil)
	return status, nil
}

// authWorker returns the Status of an authorization URL. If the authorization Status is currently 'pending', authWorker attempts to
// move the authorization to the 'valid' Status.  An error is returned if the Status can't be determined.
func (service *Service) authWorker(authUrl string, key acme.AccountKey, acmeService *acme.Service) (status string, err error) {
	// PaG the authorization
	auth, err := acmeService.GetAuth(authUrl, key)
	if err != nil {
		return "", err
	}

	// return the authoization Status
	switch auth.Status {
	// try to solve a challenge if auth is pending
	case "pending":
		auth.Status, err = service.challenges.Solve(auth.Identifier, auth.Challenges, key, acmeService)
		// return error if couldn't solve
		if err != nil {
			return "", err
		}
		// if no error, auth Status should now be "valid" or "invalid"
		fallthrough

	// if the auth is in any of these Statuses, break to return Status
	case "valid", "invalid", "deactivated", "expired", "revoked":
		// break

	// if the Status is unknown or otherwise unmatched, error
	default:
		return "", errors.New("unknown authorization status")
	}

	return auth.Status, nil
}
