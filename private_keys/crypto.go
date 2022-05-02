package private_keys

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
type algorithm struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// define the known algorithm structs
// these are used to populate the 'Name' field in API data
// this does not necessarily need to be limited to supported generation algos
// the Value must be unique
// TODO: write a go test to confirm uniqueness
func listOfAlgorithms() []algorithm {
	return []algorithm{
		{
			Value: "rsa2048",
			Name:  "RSA 2048-bit",
		},
		{
			Value: "rsa3072",
			Name:  "RSA 3072-bit",
		},
		{
			Value: "rsa4096",
			Name:  "RSA 4096-bit",
		},
		{
			Value: "ecdsap256",
			Name:  "ECDSA P-256",
		},
		{
			Value: "ecdsap384",
			Name:  "ECDSA P-384",
		},
	}
}

// return an algorithm based on its value
// if not found, return a generic unknown algorithm
func algorithmByValue(dbValue string) algorithm {
	supportedAlgorithms := listOfAlgorithms()

	for i := 0; i < len(supportedAlgorithms); i++ {
		if dbValue == supportedAlgorithms[i].Value {
			return supportedAlgorithms[i]
		}
	}

	return algorithm{
		Value: "",
		Name:  "Unknown",
	}
}

// Generate a key in PEM format based on the algorith value
// The cases do not necessarily need to match listOfAlgorithms()
// This MUST be kept in sync with the front end list of generatable algos
func generatePrivateKeyPem(algorithmValue string) (string, error) {
	var privateKeyPem string
	var err error

	switch algorithmValue {
	case "rsa2048":
		privateKeyPem, err = generateRSAPrivateKeyPem(2048)
		break
	case "rsa3072":
		privateKeyPem, err = generateRSAPrivateKeyPem(3072)
		break
	case "rsa4096":
		privateKeyPem, err = generateRSAPrivateKeyPem(4096)
		break
	case "ecdsap256":
		privateKeyPem, err = generateECDSAPrivateKeyPem(elliptic.P256())
		break
	case "ecdsap384":
		privateKeyPem, err = generateECDSAPrivateKeyPem(elliptic.P384())
		break
	default:
		return "", errors.New("key generation: invalid algorithm value")
	}

	if err != nil {
		return "", err
	}
	return privateKeyPem, nil
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
