package randomness

import (
	"crypto/rand"
	"math/big"
)

// GenerateApiKey generates a cryptographically secure API key
func GenerateApiKey() (string, error) {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 48

	key := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		key[i] = chars[num.Int64()]
	}

	return string(key), nil
}
