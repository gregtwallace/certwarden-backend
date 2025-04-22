package acme

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
)

// jsonWebKey is the JWK in the ACME protectedHeader. Members MUST be in lexicographical
// order to ensure proper thumbprint generation.
type jsonWebKey struct {
	KeyType        string `json:"kty,omitempty"`
	PublicExponent string `json:"e,omitempty"`   // RSA
	Modulus        string `json:"n,omitempty"`   // RSA
	CurveName      string `json:"crv,omitempty"` // EC
	CurvePointX    string `json:"x,omitempty"`   // EC
	CurvePointY    string `json:"y,omitempty"`   // EC
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
func (jwk *jsonWebKey) encodedSHA256PublicThumbprint() (thumbprint string, err error) {
	// rfc 7638 s3.1 - construct json object containing only the required PUBLIC members
	// of a JWK representing the key and with no whitespace or line breaks before
	// or after any syntactic elements and with the	required members ordered
	// lexicographically by the Unicode	[UNICODE] code points of the member names.
	var buf bytes.Buffer
	switch jwk.KeyType {
	case "RSA":
		_, _ = buf.WriteString(`{"e":"`)
		_, _ = buf.WriteString(jwk.PublicExponent)
		_, _ = buf.WriteString(`","kty":"RSA","n":"`)
		_, _ = buf.WriteString(jwk.Modulus)
		_, _ = buf.WriteString(`"}`)
	case "EC":
		_, _ = buf.WriteString(`{"crv":"`)
		_, _ = buf.WriteString(jwk.CurveName)
		_, _ = buf.WriteString(`","kty":"EC","x":"`)
		_, _ = buf.WriteString(jwk.CurvePointX)
		_, _ = buf.WriteString(`","y":"`)
		_, _ = buf.WriteString(jwk.CurvePointY)
		_, _ = buf.WriteString(`"}`)
	default:
		return "", errors.New("acme: jwk thumbprint: unsupported private key type")
	}

	// calculare the hash for the JSON object
	sum256 := sha256.Sum256(buf.Bytes())

	// encode as base64url
	thumbprint = encodeString(sum256[:])

	return thumbprint, nil
}
