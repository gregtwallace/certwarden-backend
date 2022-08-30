package acme

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log"
)

// jsonWebKey is the JWK in the ACME protectedHeader. Members MUST be in lexicographical
// order to ensure proper thumbprint generation.
type jsonWebKey struct {
	CurveName      string `json:"crv,omitempty"` // EC
	PublicExponent string `json:"e,omitempty"`   // RSA
	KeyType        string `json:"kty,omitempty"`
	Modulus        string `json:"n,omitempty"` // RSA
	CurvePointX    string `json:"x,omitempty"` // EC
	CurvePointY    string `json:"y,omitempty"` // EC
}

// jwk return a jwk for the AccountKey
func (accountKey *AccountKey) jwk() (jwk *jsonWebKey, err error) {
	jwk = new(jsonWebKey)

	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		jwk.KeyType = "RSA"

		jwk.PublicExponent, err = encodeInt(privateKey.E)
		if err != nil {
			return nil, err
		}
		keyBitSize := privateKey.N.BitLen()
		jwk.Modulus = encodeBigInt(privateKey.N, keyBitSize)

		return jwk, nil

	case *ecdsa.PrivateKey:
		jwk.KeyType = "EC"

		jwk.CurveName = privateKey.Curve.Params().Name

		keyBitSize := privateKey.Curve.Params().BitSize
		jwk.CurvePointX = encodeBigInt(privateKey.X, keyBitSize)
		jwk.CurvePointY = encodeBigInt(privateKey.Y, keyBitSize)

		return jwk, nil

	default:
		// break to final error return
	}

	return nil, errors.New("acme: jwk: unsupported private key type")
}

// jwkThumbprint returns the SHA-256 thumbprint for the JWK. This is calculated
// as specified in RFC7638, section 3. RFC8555 8.1 requires this for responding
// to challenges.
func (jwk *jsonWebKey) encodedThumbprint() (thumbprint string, err error) {
	// marshal the jwk
	octets, err := json.Marshal(jwk)
	if err != nil {
		return "", err
	}

	// TODO: Remove
	log.Println(string(octets))

	// calculare the hash for the JSON object
	sum256 := sha256.Sum256(octets)

	// encode as base64url
	thumbprint = encodeString(sum256[:])

	return thumbprint, nil
}
