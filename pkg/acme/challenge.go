package acme

import (
	"encoding/json"
	"net/http"
)

// ACME challenge object
type Challenge struct {
	Type      string         `json:"type"`
	Url       string         `json:"url"`
	Status    string         `json:"status"`
	Validated acmeTimeString `json:"validated,omitempty"`
	Token     string         `json:"token"`
	Error     *Error         `json:"error,omitempty"`
}

// Account response decoder
func unmarshalChallenge(bodyBytes []byte, headers http.Header) (response Challenge, err error) {
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return Challenge{}, err
	}

	return response, nil
}

// NewOrder posts a secure message to the NewOrder URL of the directory
func (service *Service) ValidateChallenge(challengeUrl string, accountKey AccountKey) (response Challenge, err error) {

	// post challenge with {} as payload signals the challenge is ready for validation
	bodyBytes, headers, err := service.postToUrlSigned(struct{}{}, challengeUrl, accountKey)
	if err != nil {
		return Challenge{}, err
	}

	// unmarshal response
	response, err = unmarshalChallenge(bodyBytes, headers)
	if err != nil {
		return Challenge{}, err
	}

	return response, nil
}

// GetChallenge does a POST-as-GET to fetch the current state of the given challenge URL
func (service *Service) GetChallenge(challengeUrl string, key AccountKey) (response Challenge, err error) {
	// POST-as-GET
	bodyBytes, headers, err := service.postAsGet(challengeUrl, key)
	if err != nil {
		return Challenge{}, err
	}

	// unmarshal response
	response, err = unmarshalChallenge(bodyBytes, headers)
	if err != nil {
		return Challenge{}, err
	}

	return response, nil
}
