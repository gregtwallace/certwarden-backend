package acme

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"
)

// encodeString returns an encoded string using the type of encoding
// ACME requires (base64 RawURL format)
func encodeString(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// encodeBinaryString returns an encoded binary string that is suitable
// for ACME
func encodeBinaryString(val int) string {
	return encodeString([]byte(strconv.FormatInt(int64(val), 2)))
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

// leToUnixTime turns an acme response formatted time into a unix time int
func leToUnixTime(leTime string) (int64, error) {
	layout := "2006-01-02T15:04:05Z"

	time, err := time.Parse(layout, leTime)
	if err != nil {
		return 0, err
	}

	return time.Unix(), nil
}
