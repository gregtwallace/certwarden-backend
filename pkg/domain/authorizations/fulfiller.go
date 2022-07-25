package authorizations

import (
	"errors"
	"fmt"
	"legocerthub-backend/pkg/acme"
	"sync"
	"time"
)

// FulfillAuthz checks each auth for validity. It returns an error if any of the authz have a problem and
// nil if all the authz are valid or were updated from pending to valid.
func (service *Service) FulfillAuthz(authUrls []string, challType string, key acme.AccountKey, isStaging bool) (err error) {
	// aysnc checking the authz for validity
	var wg sync.WaitGroup
	wgSize := len(authUrls)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	// fulfill each auth concurrently
	// TODO: Add context to cancel everything if any auth fails / invalid
	for i := range authUrls {
		go func(authUrl string, challType string, key acme.AccountKey, isStaging bool) {
			defer wg.Done()
			service.logger.Debug(authUrl)
			wgErrors <- service.fulfillAuth(authUrl, challType, key, isStaging)
		}(authUrls[i], challType, key, isStaging)
	}

	// wait for all auths to do their thing
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err = range wgErrors {
		if err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (service *Service) fulfillAuth(authUrl string, challType string, key acme.AccountKey, isStaging bool) (err error) {
	auth := new(acme.Authorization)

	// PaG the authorization
	if isStaging {
		*auth, err = service.acmeStaging.GetAuth(authUrl, key)
	} else {
		*auth, err = service.acmeStaging.GetAuth(authUrl, key)
	}
	if err != nil {
		return err
	}

	// Only proceed with negotiating auth if the auth is pending. If it is in a bad state or
	// unknown, return error. If already valid, do nothing and return no error.
	switch auth.Status {
	case "pending":
		// continue
	case "valid":
		return nil
	case "invalid", "deactivated", "expired", "revoked":
		return errors.New(fmt.Sprintf("bad authorization status: %s", auth.Status))
	default:
		return errors.New("unknown authorization status")
	}

	// TODO: Challenge solving goes here
	// move this into challenges pkg
	var chall acme.Challenge

	if isStaging {

		service.http01.AddToken(auth.Challenges[0].Token, key)

		chall, err = service.acmeStaging.ValidateChallenge(auth.Challenges[0].Url, key)

		time.Sleep(30 * time.Second)

		chall, err = service.acmeStaging.GetChallenge(auth.Challenges[0].Url, key)
		service.logger.Debug(chall)

	} else {

	}
	//

	// TODO: Confirm auth is now in 'valid' state

	return nil
}
