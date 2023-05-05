package acme

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
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

// encodeInt returns the value of an int properly encoded for ACME jwk
func encodeInt(integer int) (string, error) {
	bytesBuf := new(bytes.Buffer)

	// uint32 also seems to work, but uint does not
	err := binary.Write(bytesBuf, binary.BigEndian, uint64(integer))
	if err != nil {
		return "", err
	}

	// trim off left 00s
	// fixes: https://github.com/gregtwallace/legocerthub-backend/issues/1
	trimmedInt := bytes.TrimLeft(bytesBuf.Bytes(), "\x00")

	return encodeString(trimmedInt), nil
}

// encodeBigInt returns the bytes of a bigInt properly encoded (based on the
// bit size of the private key) for ACME jwk
func encodeBigInt(bigInt *big.Int, keyBitSize int) (encodedProp string) {
	// make buffer based on octet length (essentially divide by 8 and round up)
	octetLen := (keyBitSize + 7) >> 3
	bytesBuf := make([]byte, octetLen)

	return encodeString(bigInt.FillBytes(bytesBuf))
}

// acmeTimeString is a string that has a method to convert to unix time
// this allows convenient conversion to unix time from various acme
// responses
type timeString string

// ToUnixTime turns an acmeTimeString (acme response formatted time string) into a
// unix time int. If nil pointer, return nil; if error, return 0.
func (ats timeString) ToUnixTime() (val int) {
	layout := "2006-01-02T15:04:05Z"

	time, err := time.Parse(layout, string(ats))
	if err != nil {
		return val
	}

	return int(time.Unix())
}
