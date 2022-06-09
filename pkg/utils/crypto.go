package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// type to hold key algorithms
type Algorithm struct {
	Value             string                `json:"value"`
	Name              string                `json:"name"`
	KeyType           string                `json:"-"`
	SignatureAlg      string                `json:"-"`
	BitLen            int                   `json:"-"`
	EllipticCurveName string                `json:"-"`
	EllipticCurveFunc func() elliptic.Curve `json:"-"`
}

// define all known algorithms within their struct
// these are used to populate the 'Name' field in API data
// this does not necessarily need to be limited to supported generation algos
// the Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfAlgorithms() []Algorithm {
	return []Algorithm{
		{
			Value:        "rsa2048",
			Name:         "RSA 2048-bit",
			KeyType:      "RSA",
			SignatureAlg: "RS256",
			BitLen:       2048,
		},
		{
			Value:        "rsa3072",
			Name:         "RSA 3072-bit",
			KeyType:      "RSA",
			SignatureAlg: "", // RS384 ??
			BitLen:       3072,
		},
		{
			Value:        "rsa4096",
			Name:         "RSA 4096-bit",
			KeyType:      "RSA",
			SignatureAlg: "", // RS512 ??
			BitLen:       4096,
		},
		{
			Value:             "ecdsap256",
			Name:              "ECDSA P-256",
			KeyType:           "EC",
			SignatureAlg:      "ES256",
			EllipticCurveName: "P-256",
			EllipticCurveFunc: elliptic.P256,
		},
		{
			Value:             "ecdsap384",
			Name:              "ECDSA P-384",
			KeyType:           "EC",
			SignatureAlg:      "ES384",
			EllipticCurveName: "P-384",
			EllipticCurveFunc: elliptic.P384,
		},
	}
}

// return an algorithm based on its value
// if not found, return a generic unknown algorithm
func AlgorithmByValue(value string) Algorithm {
	supportedAlgorithms := ListOfAlgorithms()

	for i := 0; i < len(supportedAlgorithms); i++ {
		if value == supportedAlgorithms[i].Value {
			return supportedAlgorithms[i]
		}
	}

	return Algorithm{
		Value: "",
		Name:  "Unknown",
	}
}

// function to return algorithm value for an RSA algorithm of specified bits
// returns string containing the value
func rsaAlgorithmByBits(bits int) (string, error) {
	for _, item := range ListOfAlgorithms() {
		if (item.KeyType == "RSA") && (item.BitLen == bits) {
			return item.Value, nil
		}
	}
	return "", errors.New("Unsupported RSA algorithm")
}

// function to return algorithm value for an ECDSA algorithm with specified curve name
// returns string containing the value
func ecdsaAlgorithmByCurve(curveName string) (string, error) {
	for _, item := range ListOfAlgorithms() {
		if (item.KeyType == "EC") && (item.EllipticCurveName == curveName) {
			return item.Value, nil
		}
	}
	return "", errors.New("Unsupported ECDSA algorithm")
}

// Generate a key in PEM format based on the algorith value
// The cases do not necessarily need to match listOfAlgorithms()
// This MUST be kept in sync with the front end list of generatable algos
func GeneratePrivateKeyPem(algorithmValue string) (string, error) {
	algorithm := AlgorithmByValue(algorithmValue)

	if algorithm.KeyType == "RSA" {
		return generateRSAPrivateKeyPem(algorithm.BitLen)
	} else if algorithm.KeyType == "EC" {
		return generateECDSAPrivateKeyPem(algorithm.EllipticCurveFunc())
	}
	return "", errors.New("key generation: invalid algorithm value")
}

// Generate an RSA key of specified number of bits
// return a string consisting of the key in PKCS1/PEM format
func generateRSAPrivateKeyPem(bits int) (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privateKeyPem := pem.EncodeToMemory(privateKeyBlock)

	return string(privateKeyPem), nil
}

// Generate an ECDSA key of specified elliptic.Curve
// return a string consisting of the key in PKCS1/PEM format
func generateECDSAPrivateKeyPem(curve elliptic.Curve) (string, error) {
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", err
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	privateKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privateKeyPem := pem.EncodeToMemory(privateKeyBlock)

	return string(privateKeyPem), nil
}

// Parses a pemBlock to peform basic sanity check and determine key type and bit length, and
// returns a string with the algorithmValue, and
//		an error, if there is one
func PrivateKeyAlgorithmValue(pemBlock *pem.Block) (string, error) {
	switch pemBlock.Type {
	case "RSA PRIVATE KEY": // PKCS1
		privateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return "", err
		}

		// basic sanity check
		err = privateKey.Validate()
		if err != nil {
			return "", err
		}

		// find algorithm in list of supported algorithms
		algorithmValue, err := rsaAlgorithmByBits(privateKey.N.BitLen())
		if err != nil {
			return "", err
		}
		return algorithmValue, nil

	case "EC PRIVATE KEY": // SEC1, ASN.1
		privateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			return "", err
		}

		// TODO: basic sanity check?

		// find algorithm in list of supported algorithms
		algorithmValue, err := ecdsaAlgorithmByCurve(privateKey.Curve.Params().Name)
		if err != nil {
			return "", err
		}
		return algorithmValue, nil

	case "PRIVATE KEY": // PKCS8
		// TODO
		return "", errors.New("Unsupported PEM header")

	default:
		return "", errors.New("Unsupported PEM header")
	}
}

// pemDecodeAndNormalize takes a pem string and decodes it into a normalized
//  pem.Block
func PemDecodeAndNormalize(keyPem string) (*pem.Block, error) {
	// normalize line breaks and decode
	pemBlock, rest := pem.Decode(NormalizeNewLines([]byte(keyPem)))
	if pemBlock == nil {
		return nil, errors.New("Failed to decode Pem")
	}
	if len(rest) > 0 {
		return nil, errors.New("Extraneous data present")
	}

	return pemBlock, nil
}

// Decodes a PEM key string and then examines the decoded key, and
// returns a string with the same PEM key that has been sanitized, and
//		a string with the algorithmValue that identifies the key type and bit length, and
//		an error, if there is one
func ValidatePrivateKeyPem(keyPem string) (string, string, error) {
	// normalize line breaks and decode
	pemBlock, err := PemDecodeAndNormalize(keyPem)
	if err != nil {
		return "", "", err
	}

	algorithmValue, err := PrivateKeyAlgorithmValue(pemBlock)
	if err != nil {
		return "", "", err
	}

	recreatedPem := pem.EncodeToMemory(pemBlock)

	return string(recreatedPem), algorithmValue, nil
}
