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
