package key_crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// GeneratePrivateKeyPem generates a key in PEM format based on the algorithm.
// It returns an error if the algorithm is invalid or is otherwise unable to
// generate the key PEM.
func (alg Algorithm) GeneratePrivateKeyPem() (string, error) {
	algDetails := alg.details()

	switch algDetails.keyType {
	case "RSA":
		return generateRSAPrivateKeyPem(algDetails.bitLen)
	case "EC":
		return generateECDSAPrivateKeyPem(algDetails.ellipticCurveFunc())
	default:
		// break to error return
	}

	return "", errUnsupportedAlgorithm
}

// generateRSAPrivateKeyPem generates an RSA key using the specified bit length
// and returns the key in PKCS1/PEM format
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

// generateECDSAPrivateKeyPem generates an ECDSA key using the provided elliptic.Curve
// and returns the key in SEC1/PEM format
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
