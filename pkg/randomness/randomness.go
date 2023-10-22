package randomness

import (
	crypto_rand "crypto/rand"
	"math/big"
	math_rand "math/rand"
)

// lengths
const (
	lengthApiKey        = 32
	lengthFrontendNonce = 26
	lengthHexSecret     = 64
)

// character sets
const (
	charSetNumbersAndLetters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	charSetHex               = "0123456789abcdef"
)

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
// sufficiently secure entropy.
func GenerateFrontendNonce() ([]byte, error) {
	s, err := generateSecureRandomString(charSetNumbersAndLetters, lengthFrontendNonce)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// GenerateHexSecret generates a cryptographically secure random hex
// byte slice (particualrly useful for jwt secret)
func GenerateHexSecret() ([]byte, error) {
	hexString, err := generateSecureRandomString(charSetHex, lengthHexSecret)
	return []byte(hexString), err
}

// Insecure Randoms

// GenerateInsecureString creates a random string of length length using the char
// set 0-9, A-Z, and a-z. It is NOT cryptographically secure.
func GenerateInsecureString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charSetNumbersAndLetters[math_rand.Int63()%int64(len(charSetNumbersAndLetters))]
	}
	return string(bytes)
}

// GenerateInsecureInt creates a random int between [0, max). It is NOT
// cryptographically secure.
func GenerateInsecureInt(max int) int {
	return math_rand.Intn(max)
}
