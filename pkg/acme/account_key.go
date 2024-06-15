package acme

import (
	"crypto"
	"strings"
)

// AccountKey is the necessary account / key information for signed message generation
type AccountKey struct {
	Key crypto.PrivateKey
	Kid string
}

type KeyAuth string

// KeyAuthorization uses the AccountKey to create the Key Authorization for a given
// challenge token
func (accountKey *AccountKey) KeyAuthorization(token string) (keyAuth KeyAuth, err error) {
	// get jwk
	jwk, err := accountKey.jwk()
	if err != nil {
		return "", err
	}

	// calc encoded thumbprint of jwk
	encodedThumbprint, err := jwk.encodedSHA256Thumbprint()
	if err != nil {
		return "", err
	}

	return KeyAuth(strings.Join([]string{token, encodedThumbprint}, ".")), nil
}
