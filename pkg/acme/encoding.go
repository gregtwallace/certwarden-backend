package acme

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
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

// encodeInt returns the value of an int properly encoded for ACME jwk
func encodeInt(integer int) (string, error) {
	// uint32 also seems to work, but uint does not
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(integer))

	// trim off left 00s
	// fixes: https://github.com/gregtwallace/certwarden-backend/issues/1
	b = bytes.TrimLeft(b, "\x00")

	return encodeString(b), nil
}

// encodeBigInt returns the bytes of a bigInt properly encoded (based on the
// bit size of the private key) for ACME jwk
func encodeBigInt(bigInt *big.Int, keyBitSize int) (encodedProp string) {
	// make buffer based on octet length (essentially divide by 8 and round up)
	octetLen := (keyBitSize + 7) >> 3
	bytesBuf := make([]byte, octetLen)

	return encodeString(bigInt.FillBytes(bytesBuf))
}
