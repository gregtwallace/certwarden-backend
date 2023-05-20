package acme

// makeExternalAccountBinding creates a signed EAB for a newAccount using the provided params.
func (accountKey *AccountKey) makeExternalAccountBinding(eabKid, eabHmacKey, url string) (*acmeSignedMessage, error) {
	eab := new(acmeSignedMessage)
	var err error

	// make EAB protected header
	// hardcode to newAccount url
	eabHeader := protectedHeader{
		Algorithm: "HS256",
		KeyId:     eabKid,
		Url:       url,
	}

	eab.ProtectedHeader, err = encodeJson(eabHeader)
	if err != nil {
		return nil, err
	}

	// make payload (encoded jwk)
	eabPayload, err := accountKey.jwk()
	if err != nil {
		return nil, err
	}

	eab.Payload, err = encodeJson(eabPayload)
	if err != nil {
		return nil, err
	}

	// sign EAB
	err = eab.SignEAB(eabHmacKey)
	if err != nil {
		return nil, err
	}

	// add EAB to the ACME payload
	return eab, nil
}
