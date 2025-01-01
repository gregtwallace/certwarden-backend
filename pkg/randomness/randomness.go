package randomness

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"math/big"
	math_rand "math/rand/v2"
)

// lengths
const (
	lengthAES256Key     = 32
	lengthApiKey        = 32
	lengthFrontendNonce = 26
)

// character sets
const (
	charSetBase64            = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz+/"
	charSetNumbersAndLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// GenerateRandomByteSlice populates a byte slice of length with data from
// crypto/rand
func GenerateRandomByteSlice(length int) ([]byte, error) {
	slice := make([]byte, length)

	_, err := crypto_rand.Read(slice)
	if err != nil {
		return nil, err
	}

	return slice, nil
}

// GenerateAES256KeyAsBase64RawUrl generates an AES 256 encryption key
// (a 32-byte key) and then encodes the key in Base64 Raw URL format
func GenerateAES256KeyAsBase64RawUrl() (string, error) {
	// make key
	key, err := GenerateRandomByteSlice(lengthAES256Key)
	if err != nil {
		return "", err
	}

	// return encoded key
	return base64.RawURLEncoding.EncodeToString(key), nil
}

// generateSecureRandomInt returns a uniform random value in [0, max) that
// is cryptographically random.
func generateSecureRandomInt(max int) (int, error) {
	num, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -2, err
	}

	return int(num.Int64()), nil
}

// generateSecureRandomString generates a cryptographically secure random string
// based on the specified character set and length
func generateSecureRandomString(charSet string, length int) (string, error) {
	key := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := generateSecureRandomInt(len(charSet))
		if err != nil {
			return "", err
		}
		key[i] = charSet[num]
	}

	return string(key), nil
}

// GenerateApiKey generates a cryptographically secure API key with
// sufficiently secure entropy.
func GenerateApiKey() (string, error) {
	return generateSecureRandomString(charSetNumbersAndLetters, lengthApiKey)
}

// GenerateFrontendNonce generates a cryptographically secure nonce with
// sufficiently secure entropy using the base64 character set.
func GenerateFrontendNonce() ([]byte, error) {
	s, err := generateSecureRandomString(charSetBase64, lengthFrontendNonce)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// Generate32ByteSecret generates a cryptographically secure 32 byte slice
func Generate32ByteSecret() ([]byte, error) {
	return GenerateRandomByteSlice(32)
}

// Insecure Randoms

// GenerateInsecureInt creates a random int between [0, max). It is NOT
// cryptographically secure.
func GenerateInsecureInt(max int) int {
	return math_rand.IntN(max)
}

// GenerateInsecureString creates a random string of length 'length' using the char
// set 0-9, A-Z, and a-z. It is NOT cryptographically secure.
func GenerateInsecureString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charSetNumbersAndLetters[GenerateInsecureInt(len(charSetNumbersAndLetters))]
	}
	return string(bytes)
}
