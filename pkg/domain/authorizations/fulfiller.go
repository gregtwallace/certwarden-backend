package authorizations

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"sync"
	"time"
)

// FulfillAuthz checks each auth for validity. It returns an error if any of the authz have a problem and
// nil if all the authz are valid or were updated from pending to valid.
func (service *Service) FulfillAuthz(authUrls []string, method challenges.Method, key acme.AccountKey, isStaging bool) (err error) {
	// aysnc checking the authz for validity
	var wg sync.WaitGroup
	wgSize := len(authUrls)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// fulfill each auth concurrently
	// TODO: Add context to cancel everything if any auth fails / invalid?
	for i := range authUrls {
		go func(authUrl string, method challenges.Method, key acme.AccountKey, isStaging bool) {
			defer wg.Done()
			wgErrors <- service.fulfillAuth(authUrl, method, key, isStaging)
		}(authUrls[i], method, key, isStaging)
	}

	// wait for all auths to do their thing
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err = range wgErrors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *Service) fulfillAuth(authUrl string, method challenges.Method, key acme.AccountKey, isStaging bool) (err error) {
	var auth acme.Authorization

	// use loop to retry auth fulfillment as appropriate
	// max 3 tries
	for i := 1; i <= 3; i++ {
		// PaG the authorization
		if isStaging {
			auth, err = service.acmeStaging.GetAuth(authUrl, key)
		} else {
			auth, err = service.acmeStaging.GetAuth(authUrl, key)
		}
		if err != nil {
			return err
		}

		// Only proceed with negotiating auth if the auth is pending. If it is in a bad state or
		// unknown, return error. If already valid, do nothing and return no error.
		switch auth.Status {
		case "pending":
			err = service.challenges.Solve(auth.Challenges, method, key, isStaging)
			if err != nil {
				// if solve errored, break switch so loop can try again
				// TODO: Implement exponential backoff
				time.Sleep(time.Duration(i) * 30 * time.Second)
				break
			}
			// if solve didn't error, auth is now valid
			fallthrough

		case "valid":
			return nil

		case "invalid", "deactivated", "expired", "revoked":
			return errors.New(fmt.Sprintf("bad authorization status (%s)", auth.Status))

		default:
			return errors.New("unknown authorization status")
		}
	}

	return nil
}
