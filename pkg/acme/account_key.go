package acme

import (
	"crypto"
	"crypto/sha256"
	"strings"
)

// AccountKey is the necessary account / key information for signed message generation
type AccountKey struct {
	Key crypto.PrivateKey
	Kid string
}

// KeyAuthorization uses the AccountKey to create the Key Authorization for a given
// challenge token
func (accountKey *AccountKey) keyAuthorization(token string) (keyAuth string, err error) {
	// get jwk
	jwk, err := accountKey.jwk()
	if err != nil {
		return "", err
	}

	// calc encoded thumbprint of jwk
	encodedThumbprint, err := jwk.encodedThumbprint()
	if err != nil {
		return "", err
	}

	keyAuth = strings.Join([]string{token, encodedThumbprint}, ".")

	return keyAuth, nil
}

// KeyAuthorizationSHA256 uses the AccountKey to create the Key Authorization for a given
// challenge token. It then computes the SHA-256 digest of the Key Authorization. Finally,
// the base64url encoding of the digest is returned.
func (accountKey *AccountKey) keyAuthorizationEndodedSHA256(token string) (digest string, err error) {
	// get the keyAuth
	keyAuth, err := accountKey.keyAuthorization(token)
	if err != nil {
		return "", err
	}

	// calculate digest
	keyAuthDigest := sha256.Sum256([]byte(keyAuth))

	// encode and return
	return encodeString(keyAuthDigest[:]), nil
}
