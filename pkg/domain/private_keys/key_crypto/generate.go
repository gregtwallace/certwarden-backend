package key_crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GeneratePrivateKeyPem generates a key in PEM format based on the
// algorith value.  It returns an error if the algorithm is invalid
// or it is unable to generate a key in pem format.
func GeneratePrivateKeyPem(algorithmValue string) (string, error) {
	algorithm, err := AlgorithmByValue(algorithmValue)
	if err != nil {
		return "", err
	}

	if algorithm.KeyType == "RSA" {
		return generateRSAPrivateKeyPem(algorithm.BitLen)
	} else if algorithm.KeyType == "EC" {
		return generateECDSAPrivateKeyPem(algorithm.EllipticCurveFunc())
	}

	return "", errUnsupportedAlgorithm
}

// generateRSAPrivateKeyPem generates an RSA key of specified number of bits
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

// generateECDSAPrivateKeyPem generates an ECDSA key of specified elliptic.Curve
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
