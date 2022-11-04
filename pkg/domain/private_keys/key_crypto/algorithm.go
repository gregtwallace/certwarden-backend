package key_crypto

import (
	"crypto/x509"
	"encoding/json"
	"errors"
)

var errUnsupportedAlgorithm = errors.New("unsupported algorithm")

// define Algorithm
type Algorithm int

// define available Algorithms
const (
	UnknownAlgorithm Algorithm = iota

	rsa2048
	rsa3072
	rsa4096
	ecdsap256
	ecdsap384
)

// Algorithm custom JSON Marshal (turns the Algorithm into exportable AlgorithmDetails
// for output)
func (alg Algorithm) MarshalJSON() (data []byte, err error) {
	// get alg details
	details := alg.details()

	// put the exportable details into an exportable struct
	output := struct {
		StorageValue string `json:"value"`
		Name         string `json:"name"`
	}{
		StorageValue: details.storageValue,
		Name:         details.name,
	}

	// return details marshalled
	return json.Marshal(output)
}

// custom UnmarshalJSON not needed at present

// details returns the full details for the Algorithm.
func (alg Algorithm) details() algorithmDetails {
	for i := range keyAlgorithmDetails {
		if alg == keyAlgorithmDetails[i].algorithm {
			return keyAlgorithmDetails[i]
		}
	}

	// no details exist
	return algorithmDetails{}
}

// CsrSigningAlg returns the x509.SignatureAlgorithm for the Algorithm.
func (alg Algorithm) CsrSigningAlg() x509.SignatureAlgorithm {
	return alg.details().csrSignatureAlgorithm
}

// StorageValue returns the storage value of the Algorithm
func (alg Algorithm) StorageValue() string {
	return alg.details().storageValue
}
