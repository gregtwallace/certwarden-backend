package challenges

import (
	"legocerthub-backend/pkg/acme"
	"time"
)

// Solve accepts the array of challengeUrls from an authorization and solves the specific challenge
// specified by the method.
func (service *Service) Solve(challenges []acme.Challenge, method Method, key acme.AccountKey, isStaging bool) (err error) {

	// TODO: actually implement in a sane manner
	// this is cobbled together junk just to get the job done this second
	var chall acme.Challenge

	if isStaging {

		service.http01.AddToken(challenges[0].Token, key)

		chall, err = service.acmeStaging.ValidateChallenge(challenges[0].Url, key)

		time.Sleep(20 * time.Second)

		chall, err = service.acmeStaging.GetChallenge(challenges[0].Url, key)
		service.logger.Debug(chall)

	} else {
		// TODO: Prod
	}
	//

	return nil

}
