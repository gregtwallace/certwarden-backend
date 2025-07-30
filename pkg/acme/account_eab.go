package acme

import (
	"encoding/base64"
)

// makeExternalAccountBinding creates a signed message object using the provided params
// which is used in the `externalAccountBinding` of a newAccount that requires EAB
func (accountKey *AccountKey) makeExternalAccountBinding(eabKid, encodedHmacKey, url string) (*acmeSignedMessage, error) {
	// EAB Payload is the new account's jwk
	eabPayload, err := accountKey.jwk()
	if err != nil {
		return nil, err
	}

	// decode HMAC key
	decodedKey, err := base64.RawURLEncoding.DecodeString(encodedHmacKey)
	if err != nil {
		return nil, err
	}

	// make an accountKey object for signing
	eabKey := AccountKey{
		Key: decodedKey,
		Kid: eabKid,
	}

	eabMsg, err := makeAcmeSignedMessage(eabPayload, "", url, eabKey)
	if err != nil {
		return nil, err
	}

	// return EAB message
	return eabMsg, nil
}
