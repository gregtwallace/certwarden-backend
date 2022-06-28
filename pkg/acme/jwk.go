package acme

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
)

// jwk return a jwk for the AccountKey
func (accountKey *AccountKey) jwk() (jwk *jsonWebKey, err error) {
	jwk = new(jsonWebKey)

	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		jwk.KeyType = "RSA"

		jwk.PublicExponent = encodeBinaryString(privateKey.E)

		jwk.Modulus = encodeString(privateKey.N.Bytes())

		return jwk, nil

	case *ecdsa.PrivateKey:
		jwk.KeyType = "EC"

		jwk.CurveName = privateKey.Curve.Params().Name

		jwk.CurvePointX = encodeString(privateKey.X.Bytes())
		jwk.CurvePointY = encodeString(privateKey.Y.Bytes())

		return jwk, nil

	default:
		// break to final error return
	}

	return nil, errors.New("acme: jwk: unsupported private key type")
}
