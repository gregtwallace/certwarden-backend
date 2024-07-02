package key_crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"unicode"
)

var (
	errUnsupportedPem    = errors.New("unsupported pem input")
	errMismatchAlgorithm = errors.New("algorithm and pem mismatch")
)

// ValidateKeyPem sanitizes a pem string and then returns the sanitized pem string
// and algorithm. An error is returned if there is an issue with the pem provided.
// This function is used to verify keys before saving to storage. It ensures the same
// key is always formatted in the exact same manner to prevent adding duplicate keys
// to storage (i.e. since formatting is exact, the db unique constraint should always
// catch duplicates)
func ValidateAndStandardizeKeyPem(rawKeyPem string) (standardizedKeyPem string, alg Algorithm, err error) {
	// remove leading and trailing whitespace
	rawKeyPem = strings.TrimSpace(rawKeyPem)

	// check for exactly one beginning clause
	count := strings.Count(rawKeyPem, "-----BEGIN")
	if count != 1 {
		return "", UnknownAlgorithm, errUnsupportedPem
	}

	// check for exactly one ending clause
	count = strings.Count(rawKeyPem, "-----END")
	if count != 1 {
		return "", UnknownAlgorithm, errUnsupportedPem
	}

	// error if pem does not start exactly as required
	valid := strings.HasPrefix(rawKeyPem, "-----BEGIN")
	if !valid {
		return "", UnknownAlgorithm, errUnsupportedPem
	}

	// error if pem does not end exactly as required
	valid = strings.HasSuffix(rawKeyPem, "KEY-----")
	if !valid {
		return "", UnknownAlgorithm, errUnsupportedPem
	}

	// split pem into beginning, content, and end to sanitize the content
	split := strings.SplitN(rawKeyPem, "PRIVATE KEY-----", 2)
	begin := split[0] + "PRIVATE KEY-----"

	split = strings.SplitN(split[1], "-----END", 2)
	end := "-----END" + split[1]

	content := split[0]

	// properly format the PEM content, using LF for line breaks and discarding
	// any extra/other space or line break characters
	// rfc7468 s.2 specifies "parsers SHOULD ignore whitespace and other non-
	// base64 characters and MUST handle different newline conventions."
	// This process WILL ignore whitespace characters as part of the standardization
	// process, but WILL NOT ignore non-base64 characters.
	standardizedContent := []rune{}
	lineRuneCount := 0
	for _, r := range content {
		// drop any pre-existing space chars in content (standardized spaces are added below)
		if unicode.IsSpace(r) {
			continue
		}

		// rfc7468 s.2 - wrap after every 64th char
		// if current line is at 64, add standardized LF (unix new line) and reset line rune counter
		// before appending the rune
		if lineRuneCount == 64 {
			standardizedContent = append(standardizedContent, rune(10))
			lineRuneCount = 0
		}

		// append the rune to standardizedContent and increment line rune counter
		standardizedContent = append(standardizedContent, r)
		lineRuneCount++
	}

	// reassemble pem in standardized format
	// begin, LF, content, LF, end, LF
	standardizedKeyPem = begin + string(byte(10)) + string(standardizedContent) + string(byte(10)) + end + string(byte(10))

	// get the algorithm value of the new key & confirm it is supported
	_, alg, err = pemStringDecode(standardizedKeyPem, UnknownAlgorithm)
	if err != nil {
		return "", UnknownAlgorithm, err
	}

	return standardizedKeyPem, alg, nil
}

// PemStringToKey returns the PrivateKey for a given pem string
// it also verifies that the pem string is of the specified algorithm
// type, or it will return an error.
func PemStringToKey(keyPem string, alg Algorithm) (crypto.PrivateKey, error) {
	// translate pem to private key and verify that key pem is of the specified algorithm
	privateKey, _, err := pemStringDecode(keyPem, alg)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// pemStringDecode returns a crypto.PrivateKey after parsing the key pem string.
// It also determines the algorithm used in the key pem string and if alg is specified
// as something other than UnknownAlgorithm, the function will confirm alg matches the
// pem string.
// This is not intended to be called directly, but by helper functions
func pemStringDecode(keyPem string, alg Algorithm) (privKey crypto.PrivateKey, identifiedAlg Algorithm, err error) {
	// decode
	pemBlock, _ := pem.Decode([]byte(keyPem))
	if pemBlock == nil {
		return "", UnknownAlgorithm, errUnsupportedPem
	}

	// parsing depends on block type
	switch pemBlock.Type {
	case "RSA PRIVATE KEY": // PKCS1
		var rsaKey *rsa.PrivateKey
		rsaKey, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, UnknownAlgorithm, err
		}

		// basic sanity check
		err = rsaKey.Validate()
		if err != nil {
			return nil, UnknownAlgorithm, err
		}

		// find algorithm in list of supported algorithms
		identifiedAlg = rsaAlgorithmByBits(rsaKey.N.BitLen())
		if identifiedAlg == UnknownAlgorithm {
			return nil, UnknownAlgorithm, err
		}

		// success!
		privKey = rsaKey

	case "EC PRIVATE KEY": // SEC1, ASN.1
		var ecdKey *ecdsa.PrivateKey
		ecdKey, err = x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, UnknownAlgorithm, err
		}

		// find algorithm in list of supported algorithms
		identifiedAlg = ecdsaAlgorithmByCurve(ecdKey.Curve.Params().Name)
		if identifiedAlg == UnknownAlgorithm {
			return nil, UnknownAlgorithm, err
		}

		// success!
		privKey = ecdKey

	case "PRIVATE KEY": // PKCS8
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, UnknownAlgorithm, err
		}

		switch pkcs8Key := pkcs8Key.(type) {
		case *rsa.PrivateKey:
			// basic sanity check
			err = pkcs8Key.Validate()
			if err != nil {
				return nil, UnknownAlgorithm, err
			}

			// find algorithm in list of supported algorithms
			identifiedAlg = rsaAlgorithmByBits(pkcs8Key.N.BitLen())
			if identifiedAlg == UnknownAlgorithm {
				return nil, UnknownAlgorithm, err
			}

			// success!
			privKey = pkcs8Key

		case *ecdsa.PrivateKey:
			// find algorithm in list of supported algorithms
			identifiedAlg = ecdsaAlgorithmByCurve(pkcs8Key.Curve.Params().Name)
			if identifiedAlg == UnknownAlgorithm {
				return nil, UnknownAlgorithm, err
			}

			// success!
			privKey = pkcs8Key

		default:
			return nil, UnknownAlgorithm, errUnsupportedPem
		}

	default:
		return nil, UnknownAlgorithm, errUnsupportedPem
	}

	// if an alg was specified in function call, verify the pem matches
	if alg != UnknownAlgorithm && alg != identifiedAlg {
		return nil, UnknownAlgorithm, errMismatchAlgorithm
	}

	return privKey, identifiedAlg, nil
}
