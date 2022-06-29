package acme

import (
	"crypto"
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
		// all rsa use RS256
		return "RS256", nil

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
		// all rsa use RS256
		hash := crypto.SHA256
		hashed256 := sha256.Sum256([]byte(toSign))
		hashed := hashed256[:]

		// sign using the key
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
		if err != nil {
			return "", err
		}

		// make signature in the format ACME expects
		// TODO: Shouldn't need bitSize adjustment (see code in EC signature)
		encodedSignature := encodeString(signature)

		return encodedSignature, nil

	case *ecdsa.PrivateKey:
		// hash has to be generated based on the header.Algorithm or will error
		var hashed []byte
		switch privateKey.PublicKey.Params().BitSize {
		case 256:
			hashed256 := sha256.Sum256([]byte(toSign))
			hashed = hashed256[:]

		case 384:
			hashed384 := sha512.Sum384([]byte(toSign))
			hashed = hashed384[:]

		default:
			return "", errors.New("acme: failed to sign (unsupported ec bit size)")
		}

		// sign using the key
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed)
		if err != nil {
			return "", err
		}

		// make signature in the format ACME expects
		signature := append(r.Bytes(), s.Bytes()...)
		encodedSignature := encodeString(signature)

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

		return encodedSignature, nil

	default:
		// break to final error return
	}

	return "", errors.New("acme: sign: unsupported private key type")
}
