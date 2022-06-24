package acme

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"strings"
)

// signingAlg returns the proper signature algorithm based on the private key
// within an AccountKey
func (accountKey *AccountKey) signingAlg() (signatureAlgorithm string, err error) {
	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		switch privateKey.N.BitLen() {
		case 2048:
			return "RS256", nil
		case 3072:
			return "RS384", nil
		case 4096:
			return "RS512", nil
		default:
			return "", errors.New("acme: signature algorithm: unsupported rsa bit length")
		}
	case *ecdsa.PrivateKey:
		switch privateKey.Curve.Params().Name {
		case "P-256":
			return "ES256", nil
		case "P-384":
			return "ES384", nil
		default:
			return "", errors.New("acme: signature algorithm: unsupported ecdsa curve")
		}

	default:
		// break to final error return
	}

	return "", errors.New("acme: signature algorithm: unsupported private key type")
}

// Sign generates a hash for the inputted message and then signs that hash
// using the AccountKey. It returns an appropriately encoded signature string for
// ACME messages.
func (accountKey *AccountKey) Sign(message acmeSignedMessage) (string, error) {
	// create the data to sign
	toSign := strings.Join([]string{message.ProtectedHeader, message.Payload}, ".")

	// sign appropriately based on key type
	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		return "", errors.New("acme: rsa signature not implemented")

	case *ecdsa.PrivateKey:
		// hash has to be generated based on the header.Algorithm or will error
		var hash []byte
		switch privateKey.PublicKey.Params().BitSize {
		case 256:
			hash256 := sha256.Sum256([]byte(toSign))
			hash = hash256[:]

		case 384:
			hash384 := sha512.Sum384([]byte(toSign))
			hash = hash384[:]

		default:
			return "", errors.New("acme: failed to sign (unsupported ec bit size)")
		}

		// sign using the key
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
		if err != nil {
			return "", err
		}

		// make signature in the format ACME expects
		signature := encodeString(append(r.Bytes(), s.Bytes()...))

		// TODO: SHOULD NOT BE NEEDED
		// rBytes, sBytes := r.Bytes(), s.Bytes()

		// // Spec requires padding to octet length
		// // This should never be an issue with LE, but implement the spec anyway
		// octetLength := (bitSize + 7) >> 3
		// // MUST include leading zeros in the output
		// rBuf := make([]byte, octetLength-len(rBytes), octetLength)
		// sBuf := make([]byte, octetLength-len(sBytes), octetLength)

		// rBuf = append(rBuf, rBytes...)
		// sBuf = append(sBuf, sBytes...)

		// signature := encodeString(append(rBuf, sBuf...))

		return signature, nil

	default:
		// break to final error return
	}

	return "", errors.New("acme: sign: unsupported private key type")
}
