package key_crypto

import (
	"crypto/elliptic"
	"crypto/x509"
	"errors"
)

var errUnsupportedAlgorithm = errors.New("unsupported algorithm")

// Algorithm is a type to hold key algorithms
type Algorithm struct {
	Value              string                  `json:"value"`
	Name               string                  `json:"name"`
	SignatureAlgorithm x509.SignatureAlgorithm `json:"-"`
	KeyType            string                  `json:"-"` // rsa or ecdsa
	BitLen             int                     `json:"-"` // rsa
	EllipticCurveName  string                  `json:"-"` // ecdsa
	EllipticCurveFunc  func() elliptic.Curve   `json:"-"` // ecdsa
}

// ListOfAlgorithms() returns a constant list of supported algorithms
// The Value must be unique
// TODO: write a go test to confirm uniqueness
func ListOfAlgorithms() []Algorithm {
	return []Algorithm{
		{
			Value:              "rsa2048",
			Name:               "RSA 2048-bit",
			SignatureAlgorithm: x509.SHA256WithRSA,
			KeyType:            "RSA",
			BitLen:             2048,
		},
		{
			Value:              "rsa3072",
			Name:               "RSA 3072-bit",
			SignatureAlgorithm: x509.SHA256WithRSA,
			KeyType:            "RSA",
			BitLen:             3072,
		},
		{
			Value:              "rsa4096",
			Name:               "RSA 4096-bit",
			SignatureAlgorithm: x509.SHA256WithRSA,
			KeyType:            "RSA",
			BitLen:             4096,
		},
		{
			Value:              "ecdsap256",
			Name:               "ECDSA P-256",
			SignatureAlgorithm: x509.ECDSAWithSHA256,
			KeyType:            "EC",
			EllipticCurveName:  "P-256",
			EllipticCurveFunc:  elliptic.P256,
		},
		{
			Value:              "ecdsap384",
			Name:               "ECDSA P-384",
			SignatureAlgorithm: x509.ECDSAWithSHA384,
			KeyType:            "EC",
			EllipticCurveName:  "P-384",
			EllipticCurveFunc:  elliptic.P384,
		},
	}
}

// AlgorithmByValue returns an algorithm based on its Value
// Returns an error if the algorithm is not supported
func AlgorithmByValue(value string) (Algorithm, error) {
	allAlgs := ListOfAlgorithms()

	for i := range allAlgs {
		if value == allAlgs[i].Value {
			return allAlgs[i], nil
		}
	}

	return Algorithm{}, errUnsupportedAlgorithm
}

// function to return algorithm value for an RSA algorithm of specified bits
// returns string containing the value
func rsaAlgorithmByBits(bits int) (string, error) {
	allAlgs := ListOfAlgorithms()

	for i := range allAlgs {
		if (allAlgs[i].KeyType == "RSA") && (allAlgs[i].BitLen == bits) {
			return allAlgs[i].Value, nil
		}
	}
	return "", errUnsupportedAlgorithm
}

// function to return algorithm value for an ECDSA algorithm with specified curve name
// returns string containing the value
func ecdsaAlgorithmByCurve(curveName string) (string, error) {
	allAlgs := ListOfAlgorithms()

	for i := range allAlgs {
		if (allAlgs[i].KeyType == "EC") && (allAlgs[i].EllipticCurveName == curveName) {
			return allAlgs[i].Value, nil
		}
	}
	return "", errUnsupportedAlgorithm
}
