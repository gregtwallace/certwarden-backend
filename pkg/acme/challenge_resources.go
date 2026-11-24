package acme

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

var (
	ErrChallengeTypeNotFound           = errors.New("challenge type not found in challenges array")
	ErrChallengeMalformed              = errors.New("challenge malformed")
	ErrChallengeTypeDoesntSupportCNAME = errors.New("challenge type doesnt support CNAME use")
)

// ValidationResourceDns01 returns the dnsRecord name and value to provision
// in response to a Dns01 challenge for a given domain and keyAuth
func ValidationResourceDns01(domain string, keyAuth KeyAuth) (dnsRecordName string, dnsRecordValue string) {
	// dns record name is just the domain prepended with the special acme prefix
	dnsRecordName = "_acme-challenge." + domain

	// dns record value is the base64 encoded sha256 of key authorization
	// calculate digest
	keyAuthDigest := sha256.Sum256([]byte(keyAuth))

	// encode it
	dnsRecordValue = encodeString(keyAuthDigest[:])

	return dnsRecordName, dnsRecordValue
}

// SelectChallenge returns the challenge of the specified type. If required by spec, some
// properties of the challenge are validated.
// If the specified type is not found or if the challenge is invalid, an error is returned.
func SelectChallenge(challengeType ChallengeType, challenges []Challenge) (Challenge, error) {
	// range to the correct challenge to solve based on ACME Challenge Type (from provider)
	var challenge Challenge
	found := false

	for i := range challenges {
		if challenges[i].Type == challengeType {
			found = true
			challenge = challenges[i]
			break
		}
	}
	if !found {
		return Challenge{}, ErrChallengeTypeNotFound
	}

	// some challenge types have constraints
	if challengeType == ChallengeTypeDnsPersist01 {
		if len(challenge.IssuerDomainNames) == 0 || len(challenge.IssuerDomainNames) > 10 {
			return Challenge{}, fmt.Errorf("challenges: dns-persist-01 challenge had invalid issuer domain name count of %d (%w)",
				len(challenge.IssuerDomainNames), ErrChallengeMalformed)
		}
	}

	return challenge, nil
}

// DNSChallengeCNAMEInfo returns the from and to domains for a CNAME record, if one is needed
func DNSChallengeCNAMEInfo(dnsIdValue, provisionDomain string, challengeType ChallengeType) (from, to string, _ error) {
	// no CNAME needed
	if provisionDomain == dnsIdValue {
		return "", "", nil
	}

	// exact cname domain depends on challenge type
	switch challengeType {
	case ChallengeTypeDns01:
		from = "_acme-challenge." + dnsIdValue
		to = "_acme-challenge." + provisionDomain

	case ChallengeTypeDnsPersist01:
		from = "_validation-persist." + dnsIdValue
		to = "_validation-persist." + provisionDomain

	case ChallengeTypeHttp01:
		from = dnsIdValue
		to = provisionDomain

	default:
		return "", "", ErrChallengeTypeDoesntSupportCNAME
	}

	return from, to, nil
}
