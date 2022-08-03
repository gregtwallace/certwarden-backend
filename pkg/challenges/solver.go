package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

var ErrChallengeNotValid = errors.New("challenge is not valid")

// Solve accepts the array of challengeUrls from an authorization and solves the specific challenge
// specified by the method.
func (service *Service) Solve(challenges []acme.Challenge, method Method, key acme.AccountKey, isStaging bool) (err error) {

	// TODO: actually implement in a sane manner
	// this is cobbled together junk just to get the job done this second
	var chall acme.Challenge

	if isStaging {

		service.http01.AddToken(challenges[0].Token, key)

		time.Sleep(10 * time.Second) // TODO: Remove - artificial delay for testing duplicative stuff

		chall, err = service.acmeStaging.ValidateChallenge(challenges[0].Url, key)

		time.Sleep(20 * time.Second)

		chall, err = service.acmeStaging.GetChallenge(challenges[0].Url, key)
		service.logger.Debug(chall)

	} else {
		// TODO: Prod
	}
	//

	if chall.Status != "valid" {
		return ErrChallengeNotValid
	}

	return nil

}
