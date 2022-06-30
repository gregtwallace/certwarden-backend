package acme

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"time"
)

// encodeString returns an encoded string using the type of encoding
// ACME requires (base64 RawURL format)
func encodeString(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// encodeJson transforms a data object into json and then encodes it
// in the required string format
func encodeJson(data any) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return encodeString(jsonBytes), nil
}

// encodeRsaExponent returns the value of e (rsa private key's exponent),
// properly encoded for ACME jwk
func encodeRsaExponent(privateKey rsa.PrivateKey) (string, error) {
	bytesBuf := new(bytes.Buffer)
	var err error

	// uint32 also seems to work, but uint does not
	err = binary.Write(bytesBuf, binary.BigEndian, uint64(privateKey.E))
	if err != nil {
		return "", err
	}

	return encodeString(bytesBuf.Bytes()), nil
}

// encodeRsaModulus returns the value of n (rsa private key's modulus),
// properly encoded for ACME jwk
func encodeRsaModulus(privateKey rsa.PrivateKey) (string, error) {
	bytesBuf := new(bytes.Buffer)
	var err error

	err = binary.Write(bytesBuf, binary.BigEndian, privateKey.N.Bytes())
	if err != nil {
		return "", err
	}

	return encodeString(bytesBuf.Bytes()), nil
}

// acmeToUnixTime turns an acme response formatted time into a unix time int
func acmeToUnixTime(acmeTime string) (int, error) {
	layout := "2006-01-02T15:04:05Z"

	time, err := time.Parse(layout, acmeTime)
	if err != nil {
		return 0, err
	}

	return int(time.Unix()), nil
}
