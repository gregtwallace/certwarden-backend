package key_crypto

import (
	"crypto/elliptic"
	"errors"
)

// Algorithm is a type to hold key algorithms
type Algorithm struct {
	Value             string                `json:"value"`
	Name              string                `json:"name"`
	KeyType           string                `json:"-"` // rsa or ecdsa
	BitLen            int                   `json:"-"` // rsa
	EllipticCurveName string                `json:"-"` // ecdsa
	EllipticCurveFunc func() elliptic.Curve `json:"-"` // ecdsa
}

// ListOfAlgorithms() returns a constant list of supported algorithms
// The Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfAlgorithms() []Algorithm {
	return []Algorithm{
		{
			Value:   "rsa2048",
			Name:    "RSA 2048-bit",
			KeyType: "RSA",
			BitLen:  2048,
		},
		{
			Value:   "rsa3072",
			Name:    "RSA 3072-bit",
			KeyType: "RSA",
			BitLen:  3072,
		},
		{
			Value:   "rsa4096",
			Name:    "RSA 4096-bit",
			KeyType: "RSA",
			BitLen:  4096,
		},
		{
			Value:             "ecdsap256",
			Name:              "ECDSA P-256",
			KeyType:           "EC",
			EllipticCurveName: "P-256",
			EllipticCurveFunc: elliptic.P256,
		},
		{
			Value:             "ecdsap384",
			Name:              "ECDSA P-384",
			KeyType:           "EC",
			EllipticCurveName: "P-384",
			EllipticCurveFunc: elliptic.P384,
		},
	}
}

// AlgorithmByValue returns an algorithm based on its Value
// Returns an error if the algorithm is not supported
func AlgorithmByValue(value string) (Algorithm, error) {
	// TODO: Rework using range
	supportedAlgorithms := ListOfAlgorithms()

	for i := 0; i < len(supportedAlgorithms); i++ {
		if value == supportedAlgorithms[i].Value {
			return supportedAlgorithms[i], nil
		}
	}

	return Algorithm{}, errors.New("key_crypto: unsupported algorithm")
}

// function to return algorithm value for an RSA algorithm of specified bits
// returns string containing the value
func rsaAlgorithmByBits(bits int) (string, error) {
	for _, item := range ListOfAlgorithms() {
		if (item.KeyType == "RSA") && (item.BitLen == bits) {
			return item.Value, nil
		}
	}
	return "", errors.New("key_crypto: unsupported rsa algorithm")
}

// function to return algorithm value for an ECDSA algorithm with specified curve name
// returns string containing the value
func ecdsaAlgorithmByCurve(curveName string) (string, error) {
	for _, item := range ListOfAlgorithms() {
		if (item.KeyType == "EC") && (item.EllipticCurveName == curveName) {
			return item.Value, nil
		}
	}
	return "", errors.New("key_crypto: unsupported ecdsa algorithm")
}
