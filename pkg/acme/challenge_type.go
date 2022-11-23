package acme

import (
	"errors"
)

var (
	errUnsupportedChallengeType = errors.New("unsupported challenge type")
	errWrongIdentifierType      = errors.New("acme identifier is not of type dns")
)

// Define challenge types (per RFC 8555)
type ChallengeType string

const (
	UnknownChallengeType ChallengeType = ""

	ChallengeTypeHttp01 = "http-01"
	ChallengeTypeDns01  = "dns-01"
)

// ValidationResource creates the resource name and content that are required
// to succesfully validate an ACME Challenge.
func (challType ChallengeType) ValidationResource(identifier Identifier, key AccountKey, token string) (name string, content string, err error) {
	// resource name
	name, err = challType.ValidationResourceName(identifier, token)
	if err != nil {
		return "", "", err
	}

	// resource content
	content, err = challType.validationResourceContent(identifier, key, token)
	if err != nil {
		return "", "", err
	}

	return name, content, nil
}

// ValidationResourceName returns the resource name that is required to
// validate the specified identifier
func (challType ChallengeType) ValidationResourceName(identifier Identifier, token string) (name string, err error) {
	// verify identifier is the proper type (only dns identifiers are supported)
	if identifier.Type != identifierTypeDns {
		return "", errWrongIdentifierType
	}

	// return resource name based on challenge type
	switch challType {
	// http-01 (HTTP Challenge - RFC 8555 8.3), http-01 uses the token as
	// the resource name
	case ChallengeTypeHttp01:
		name = token

	// dns-01 (DNS Challenge - RFC 8555 8.4)
	case ChallengeTypeDns01:
		// dns-01 uses "_acme-challenge." prepended to the dns identifier value
		// (e.g. "_acme-challenge.idendifier.example.com") as the resource name
		name = "_acme-challenge." + identifier.Value

	// any other type is error
	default:
		return "", errUnsupportedChallengeType
	}

	return name, nil
}

// validationResourceContent returns the resource content that is required to
// validate the specified identifier
func (challType ChallengeType) validationResourceContent(identifier Identifier, key AccountKey, token string) (content string, err error) {
	// verify identifier is the proper type (only dns identifiers are supported)
	if identifier.Type != identifierTypeDns {
		return "", errWrongIdentifierType
	}

	// return resource info based on challenge type
	switch challType {
	// http-01 (HTTP Challenge - RFC 8555 8.3)
	// http-01 uses the keyAuth as the resource content
	case ChallengeTypeHttp01:
		content, err = key.keyAuthorization(token)

	// dns-01 (DNS Challenge - RFC 8555 8.4)
	// dns-01 uses the keyAuth's SHA-256 Encoded Hash as the resource content.
	case ChallengeTypeDns01:
		content, err = key.keyAuthorizationEndodedSHA256(token)

	// any other type is error
	default:
		return "", errUnsupportedChallengeType
	}

	// error check after trying to generate content
	if err != nil {
		return "", err
	}

	return content, nil
}
