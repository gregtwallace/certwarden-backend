package acme

import (
	"encoding/json"
)

// Define challenge types (per RFC 8555)
type ChallengeType string

const (
	UnknownChallengeType ChallengeType = ""

	ChallengeTypeHttp01 ChallengeType = "http-01"
	ChallengeTypeDns01  ChallengeType = "dns-01"
)

// ACME challenge object
type Challenge struct {
	Type      ChallengeType `json:"type"`
	Url       string        `json:"url"`
	Status    string        `json:"status"`
	Validated timeString    `json:"validated,omitempty"`
	Token     string        `json:"token"`
	Error     *Error        `json:"error,omitempty"`
}

// Account response decoder
func unmarshalChallenge(jsonResp json.RawMessage) (chall Challenge, err error) {
	err = json.Unmarshal(jsonResp, &chall)
	if err != nil {
		return Challenge{}, err
	}

	return chall, nil
}

// NewOrder posts a an empty object to the challenge URL which informs ACME that the
// challenge is ready to be validated
func (service *Service) ValidateChallenge(challengeUrl string, accountKey AccountKey) (chall Challenge, err error) {
	// post challenge with {} as payload signals the challenge is ready for validation
	jsonResp, _, err := service.postToUrlSigned(struct{}{}, challengeUrl, accountKey)
	if err != nil {
		return Challenge{}, err
	}

	// unmarshal response
	chall, err = unmarshalChallenge(jsonResp)
	if err != nil {
		return Challenge{}, err
	}

	return chall, nil
}

// GetChallenge does a POST-as-GET to fetch the current state of the given challenge URL
func (service *Service) GetChallenge(challengeUrl string, key AccountKey) (chall Challenge, err error) {
	// POST-as-GET
	jsonResp, _, err := service.postAsGet(challengeUrl, key)
	if err != nil {
		return Challenge{}, err
	}

	// unmarshal response
	chall, err = unmarshalChallenge(jsonResp)
	if err != nil {
		return Challenge{}, err
	}

	return chall, nil
}
