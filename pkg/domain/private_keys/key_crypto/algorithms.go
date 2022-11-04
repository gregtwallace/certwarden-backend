package key_crypto

import (
	"crypto/elliptic"
	"crypto/x509"
)

// Define AlgorithmDetails which contains details about the defined Algorithms.
type algorithmDetails struct {
	algorithm             Algorithm
	storageValue          string
	name                  string
	csrSignatureAlgorithm x509.SignatureAlgorithm
	keyType               string                // rsa or ecdsa
	bitLen                int                   // rsa
	ellipticCurveName     string                // ecdsa
	ellipticCurveFunc     func() elliptic.Curve // ecdsa
}

var keyAlgorithmDetails = []algorithmDetails{
	{
		algorithm:             rsa2048,
		storageValue:          "rsa2048",
		name:                  "RSA 2048-bit",
		csrSignatureAlgorithm: x509.SHA256WithRSA,
		keyType:               "RSA",
		bitLen:                2048,
	},
	{
		algorithm:             rsa3072,
		storageValue:          "rsa3072",
		name:                  "RSA 3072-bit",
		csrSignatureAlgorithm: x509.SHA256WithRSA,
		keyType:               "RSA",
		bitLen:                3072,
	},
	{
		algorithm:             rsa4096,
		storageValue:          "rsa4096",
		name:                  "RSA 4096-bit",
		csrSignatureAlgorithm: x509.SHA256WithRSA,
		keyType:               "RSA",
		bitLen:                4096,
	},
	{
		algorithm:             ecdsap256,
		storageValue:          "ecdsap256",
		name:                  "ECDSA P-256",
		csrSignatureAlgorithm: x509.ECDSAWithSHA256,
		keyType:               "EC",
		ellipticCurveName:     "P-256",
		ellipticCurveFunc:     elliptic.P256,
	},
	{
		algorithm:             ecdsap384,
		storageValue:          "ecdsap384",
		name:                  "ECDSA P-384",
		csrSignatureAlgorithm: x509.ECDSAWithSHA384,
		keyType:               "EC",
		ellipticCurveName:     "P-384",
		ellipticCurveFunc:     elliptic.P384,
	},
}

// ListOfAlgorithms() returns a slice of all Algorithms
func ListOfAlgorithms() (algs []Algorithm) {
	// loop through details to make slice of methods
	for i := range keyAlgorithmDetails {
		algs = append(algs, keyAlgorithmDetails[i].algorithm)
	}

	return algs
}

// AlgorithmByValue returns an algorithm based on its Value
func AlgorithmByStorageValue(value string) Algorithm {
	for i := range keyAlgorithmDetails {
		if value == keyAlgorithmDetails[i].storageValue {
			return keyAlgorithmDetails[i].algorithm
		}
	}

	return UnknownAlgorithm
}

// rsaAlgorithmByBits returns the Algorithm corresponding to an RSA
// key of the specified bit length.
func rsaAlgorithmByBits(bits int) Algorithm {
	for i := range keyAlgorithmDetails {
		if (keyAlgorithmDetails[i].keyType == "RSA") && (keyAlgorithmDetails[i].bitLen == bits) {
			return keyAlgorithmDetails[i].algorithm
		}
	}

	return UnknownAlgorithm
}

// ecdsaAlgorithmByCurve returns the Algorithm corresponding to an ECDSA
// key that uses the specified curve canonical name
func ecdsaAlgorithmByCurve(curveName string) Algorithm {
	for i := range keyAlgorithmDetails {
		if (keyAlgorithmDetails[i].keyType == "EC") && (keyAlgorithmDetails[i].ellipticCurveName == curveName) {
			return keyAlgorithmDetails[i].algorithm
		}
	}

	return UnknownAlgorithm
}
