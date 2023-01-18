package randomness

import (
	"crypto/rand"
	"math/big"
)

// GenerateInt returns a uniform random value in [0, max)
func GenerateRandomInt(max int) (int, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -2, err
	}

	return int(num.Int64()), nil
}

// generateRandomString generates a cryptographically secure random string
// based on the specified character set and length
func generateRandomString(charSet string, length int) (string, error) {
	key := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := GenerateRandomInt(len(charSet))
		if err != nil {
			return "", err
		}
		key[i] = charSet[num]
	}

	return string(key), nil
}

// GenerateApiKey generates a cryptographically secure API key
func GenerateApiKey() (string, error) {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 48

	return generateRandomString(chars, length)
}

// generate random hex byte slice (using as jwt secret)
func GenerateHexSecret() ([]byte, error) {
	const chars = "0123456789abcdef"
	const length = 64

	hexString, err := generateRandomString(chars, length)
	return []byte(hexString), err
}
