package acme

import (
	"time"
)

// Define challenge types (per RFC 8555)
type ChallengeType string

const (
	ChallengeTypeUnknown ChallengeType = ""

	ChallengeTypeHttp01       ChallengeType = "http-01"
	ChallengeTypeDns01        ChallengeType = "dns-01"
	ChallengeTypeDnsPersist01 ChallengeType = "dns-persist-01"
)

// ACME challenge object
type Challenge struct {
	Type      ChallengeType `json:"type"`
	Url       string        `json:"url"`
	Status    string        `json:"status"`
	Validated time.Time     `json:"validated,omitempty"`
	Token     string        `json:"token"`
	Error     *Error        `json:"error,omitempty"`

	// for dns-persist-01
	IssuerDomainNames []string `json:"issuer-domain-names,omitempty"`
}

// DoChallengeValidation posts a an empty object to the challenge URL which informs
// ACME that the challenge is ready to be validated
// Note: per rfc8555 s. 7.5.1 "The server provides a 200 (OK) response with the updated challenge
// object as its body." -- However, the updated challenge body isn't really needed for anything,
// so discard it here instead.
func (service *Service) DoChallengeValidation(challengeUrl string, accountKey AccountKey) error {
	// post challenge with {} as payload signals the challenge is ready for validation
	_, _, err := service.postToUrlSigned(struct{}{}, challengeUrl, accountKey)
	if err != nil {
		return err
	}

	return nil
}
