package acme_utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
)

func LeToUnixTime(leTime string) (int64, error) {
	layout := "2006-01-02T15:04:05Z"

	time, err := time.Parse(layout, leTime)
	if err != nil {
		return 0, err
	}

	return time.Unix(), nil
}

// encodeAcmeData transforms a data object into json and then encodes it
//  in base64 RawURL format
func encodeAcmeData(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	log.Println(string(jsonBytes))

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

// JwkEcKey creates an ACME Json Web Key (JWK) from the inputted private key
func JwkEcKey(privateKey *ecdsa.PrivateKey) *jsonWebKey {
	var jsonWebKey jsonWebKey

	jsonWebKey.KeyType = "EC"
	jsonWebKey.CurveName = privateKey.Curve.Params().Name

	xBytes := privateKey.X.Bytes()
	yBytes := privateKey.Y.Bytes()
	octetLength := (privateKey.Params().BitSize + 7) >> 3

	// JWK has to pad x and y so they are octetLength (divisible by 8)
	// should be irrelevant for this program (bits are all increments of 8, but might as well do this just in case)
	xBuf := make([]byte, octetLength-len(xBytes), octetLength)
	yBuf := make([]byte, octetLength-len(yBytes), octetLength)
	xBuf = append(xBuf, xBytes...)
	yBuf = append(yBuf, yBytes...)

	jsonWebKey.CurvePointX = base64.RawURLEncoding.EncodeToString(xBuf)
	jsonWebKey.CurvePointY = base64.RawURLEncoding.EncodeToString(yBuf)

	return &jsonWebKey
}

// TODO JWK func for RSA keys

// acmeSignature generates the encoded signature that should be loaded into message.Signature
func acmeEcSignature(message acmeMessage, privateKey *ecdsa.PrivateKey) (string, error) {
	toSign := strings.Join([]string{message.ProtectedHeader, message.Payload}, ".")
	bitSize := privateKey.PublicKey.Params().BitSize

	// hash has to be generated based on the header.Algorithm or will error
	var hash []byte
	if bitSize == 256 {
		hash256 := sha256.Sum256([]byte(toSign))
		hash = hash256[:]
	} else if bitSize == 384 {
		hash384 := sha512.Sum384([]byte(toSign))
		hash = hash384[:]
	} else {
		return "", errors.New("Unsupported bitSize")
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return "", err
	}

	rBytes, sBytes := r.Bytes(), s.Bytes()

	// Spec requires padding to octet length
	// This should never be an issue with LE, but implement the spec anyway
	octetLength := (bitSize + 7) >> 3
	// MUST include leading zeros in the output
	rBuf := make([]byte, octetLength-len(rBytes), octetLength)
	sBuf := make([]byte, octetLength-len(sBytes), octetLength)

	rBuf = append(rBuf, rBytes...)
	sBuf = append(sBuf, sBytes...)

	signature := append(rBuf, sBuf...)

	return base64.RawURLEncoding.EncodeToString(signature), nil
}

// TODO Signature func for RSA keys
