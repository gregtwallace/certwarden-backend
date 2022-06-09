package utils

import "bytes"

func NormalizeNewLines(data []byte) []byte {
	// Fix windows line breaks
	data = bytes.Replace(data, []byte{13, 10}, []byte{10}, -1)
	// Fix mac line breaks
	data = bytes.Replace(data, []byte{13}, []byte{10}, -1)

	return data
}
