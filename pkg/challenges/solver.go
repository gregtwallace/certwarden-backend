package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

var (
	errChallengeRetriesExhausted = errors.New("challenge failed (out of retries)")
	errChallengeTypeNotFound     = errors.New("intended challenge type not found")
)

// Solve accepts a slice of challenges from an authorization and solves the specific challenge
// specified by the method. Valid or invalid status is returned.  An error is returned if can't resolve
// a valid or invalid state.
func (service *Service) Solve(challenges []acme.Challenge, method Method, key acme.AccountKey, isStaging bool) (status string, err error) {
	var challenge acme.Challenge
	found := false

	// range to the correct challenge to solve based on Type
	for i := range challenges {
		if challenges[i].Type == method.Type {
			found = true
			challenge = challenges[i]
		}
	}
	if !found {
		return "", errChallengeTypeNotFound
	}

	// make keyAuth for challenge solvers
	keyAuth, err := key.KeyAuthorization(challenge.Token)
	if err != nil {
		return "", err
	}

	// solve using the proper method
	switch method.Value {
	case "http-01-internal":
		status, err = service.solveHttp01Internal(challenge, keyAuth, key, isStaging)
		if err != nil {
			return "", err
		}

	default:
		return "", errUnsupportedMethod
	}

	return status, nil
}

// solveHttp01Internal adds the token to the http01 server, validates the http challenge, and then
// removes the token.  An error is returned instead of status if the challenge doesn't ever reach
// valid or invalid state (or any other error occurs)
func (service *Service) solveHttp01Internal(challenge acme.Challenge, keyAuth string, key acme.AccountKey, isStaging bool) (status string, err error) {
	// make pointer for the correct acme.Service (to avoid repeat of if/else)
	var acmeService *acme.Service
	if isStaging {
		acmeService = service.acmeStaging
	} else {
		acmeService = service.acmeProd
	}

	// add token to internal http server and defer removal
	service.http01.AddToken(challenge.Token, keyAuth)
	defer service.http01.RemoveToken(challenge.Token)

	// TODO Remove Delay - this is to artifically allow requests to stack
	time.Sleep(10 * time.Second)

	// inform ACME that the challenge is ready
	_, err = acmeService.ValidateChallenge(challenge.Url, key)
	if err != nil {
		return "", err
	}

	// monitor for processing to complete (max 5 tries, 20 seconds apart each)
	for i := 1; i <= 5; i++ {
		// sleep to allow ACME to process
		time.Sleep(20 * time.Second)

		// get challenge and check for error or final Statuses
		challenge, err = acmeService.GetChallenge(challenge.Url, key)
		if err != nil {
			return "", err
		}

		// return Status if it has reached a final status
		if challenge.Status == "valid" {
			return challenge.Status, nil
		} else if challenge.Status == "invalid" {
			service.logger.Debug(challenge.Error)
			return challenge.Status, nil
		}
		// else repeat loop
	}

	// loop ended without reaching valid or invalid Status
	return "", errChallengeRetriesExhausted
}
