package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
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

// Sign generates a hash for the message and then signs that hash using the AccountKey.
// It modifies the message to add the signature or returns an error.
// ACME messages.
func (asm *acmeSignedMessage) Sign(accountKey AccountKey) error {
	encodedSignature := ""

	// create the data to sign
	toSign := asm.dataToSign()

	// sign appropriately based on key type
	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		// all rsa use RS256
		hash := crypto.SHA256
		hashed256 := sha256.Sum256(toSign)
		hashed := hashed256[:]

		// sign using the key
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
		if err != nil {
			return err
		}

		// Make signature in the format ACME expects. Padding should not be required
		// for RSA.
		encodedSignature = encodeString(signature)

	case *ecdsa.PrivateKey:
		// hash has to be generated based on the header.Algorithm or will error
		var hashed []byte
		bitSize := privateKey.PublicKey.Params().BitSize
		switch bitSize {
		case 256:
			hashed256 := sha256.Sum256(toSign)
			hashed = hashed256[:]

		case 384:
			hashed384 := sha512.Sum384(toSign)
			hashed = hashed384[:]

		default:
			return errors.New("acme: failed to sign (unsupported ec bit size)")
		}

		// sign using the key
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashed)
		if err != nil {
			return err
		}

		// ACME expects these values to be zero padded
		rPadded := padBytes(r.Bytes(), bitSize)
		sPadded := padBytes(s.Bytes(), bitSize)

		// combine the buffers and encode
		encodedSignature = encodeString(append(rPadded, sPadded...))

	default:
		// not supported
		return errors.New("acme: sign: unsupported private key type")
	}

	asm.Signature = encodedSignature
	return nil
}

// padBytes pads data to an appropriate byte size based on the specified
// number of bits (which generally comes from the key bit size)
func padBytes(data []byte, bitSize int) (padded []byte) {
	// calculate byte length (bits rounded up to nearest 8)
	octetLength := (bitSize + 7) >> 3

	// make new buffer of byte length
	padded = make([]byte, octetLength-len(data), octetLength)

	// insert the data into the padded buffer
	padded = append(padded, data...)

	return padded
}

// SignEAB modifies the ASM to add an HMAC signature to it using the
// provided encoded hmac key
func (asm *acmeSignedMessage) SignEAB(encodedHmacKey string) error {
	// decode key
	decodedKey, err := base64.RawURLEncoding.DecodeString(encodedHmacKey)
	if err != nil {
		return err
	}

	// mac signature
	mac := hmac.New(sha256.New, decodedKey)
	_, err = mac.Write(asm.dataToSign())
	if err != nil {
		return err
	}

	// encoded Signature
	asm.Signature = encodeString(mac.Sum(nil))

	return nil
}
