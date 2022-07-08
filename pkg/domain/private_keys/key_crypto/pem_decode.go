package key_crypto

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var (
	errUnsupportedPem    = errors.New("unsupported pem input")
	errMismatchAlgorithm = errors.New("algorithm and pem mismatch")
	errMissingAlgorithm  = errors.New("missing algorithm")
)

// ValidateKeyPem sanitizes a pem string and then returns the sanitized pem string,
// the algorithm.Value, and an error if the pem has a problem.
// This function is used to verify imported keys before saving
// to storage.
func ValidateKeyPem(keyPem string) (string, string, error) {
	// normalize line breaks
	pemBytes := []byte(keyPem)
	// windows
	pemBytes = bytes.Replace(pemBytes, []byte{13, 10}, []byte{10}, -1)
	// mac
	pemBytes = bytes.Replace(pemBytes, []byte{13}, []byte{10}, -1)

	pemNormalized := string(pemBytes)

	// get the algorithm value of the new key / confirm it is a supported
	// pkcs and algorithm type
	_, algorithmValue, err := pemStringDecode(pemNormalized, "")
	if err != nil {
		return "", "", err
	}

	return pemNormalized, algorithmValue, nil
}

// PemStringToKey returns the PrivateKey for a given pem string
// it also verifies that the pem string is of the specified algorithm
// type, or it will return an error.
func PemStringToKey(keyPem string, algorithmValue string) (crypto.PrivateKey, error) {
	// require algorithm validation
	if algorithmValue == "" {
		return nil, errMissingAlgorithm
	}

	// translate pem to private key and verify that key pem is of the specified
	// algorithm type
	privateKey, _, err := pemStringDecode(keyPem, algorithmValue)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// pemStringDecode returns a crypto.PrivateKey after parsing the pemBlock,
// it also returns the algorithm.Value and an error
// This is not intended to be called directly, but by helper functions
func pemStringDecode(keyPem string, algorithmValue string) (crypto.PrivateKey, string, error) {
	var privateKey crypto.PrivateKey
	var pemAlgorithValue string
	var err error

	// decode
	pemBlock, _ := pem.Decode([]byte(keyPem))
	if pemBlock == nil {
		return "", "", errUnsupportedPem
	}

	// parsing depends on block type
	switch pemBlock.Type {
	case "RSA PRIVATE KEY": // PKCS1
		var rsaKey *rsa.PrivateKey
		rsaKey, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, "", err
		}

		// basic sanity check
		err = rsaKey.Validate()
		if err != nil {
			return nil, "", err
		}

		// find algorithm in list of supported algorithms
		pemAlgorithValue, err = rsaAlgorithmByBits(rsaKey.N.BitLen())
		if err != nil {
			return nil, "", err
		}

		// success!
		privateKey = rsaKey

	case "EC PRIVATE KEY": // SEC1, ASN.1
		var ecdKey *ecdsa.PrivateKey
		ecdKey, err = x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, "", err
		}

		// find algorithm in list of supported algorithms
		pemAlgorithValue, err = ecdsaAlgorithmByCurve(ecdKey.Curve.Params().Name)
		if err != nil {
			return nil, "", err
		}

		// success!
		privateKey = ecdKey

	case "PRIVATE KEY": // PKCS8
		pkcs8Key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, "", err
		}

		switch pkcs8Key := pkcs8Key.(type) {
		case *rsa.PrivateKey:
			// basic sanity check
			err = pkcs8Key.Validate()
			if err != nil {
				return nil, "", err
			}

			// find algorithm in list of supported algorithms
			pemAlgorithValue, err = rsaAlgorithmByBits(pkcs8Key.N.BitLen())
			if err != nil {
				return nil, "", err
			}

			// success!
			privateKey = pkcs8Key

		case *ecdsa.PrivateKey:
			// find algorithm in list of supported algorithms
			pemAlgorithValue, err = ecdsaAlgorithmByCurve(pkcs8Key.Curve.Params().Name)
			if err != nil {
				return nil, "", err
			}

			// success!
			privateKey = pkcs8Key

		default:
			return nil, "", errUnsupportedPem
		}

	default:
		return nil, "", errUnsupportedPem
	}

	// if an algorithmValue was specified in function call, verify the pem matches
	if algorithmValue != "" && algorithmValue != pemAlgorithValue {
		return nil, "", errMismatchAlgorithm
	}

	return privateKey, pemAlgorithValue, nil
}
